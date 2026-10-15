package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Meddoez/job-search-tracker/backend/internal/models"
	"github.com/gin-gonic/gin"
)

type stubProfileJobRepository struct {
	jobs   []models.SavedProfileJob
	err    error
	called bool
	id     int
	jobID  int
	status *string
	notes  *string
	result models.ProfileJob
}

func (repo *stubProfileJobRepository) SaveProfileJob(context.Context, int, models.JobSearchResult) (models.ProfileJob, error) {
	return models.ProfileJob{}, nil
}

func (repo *stubProfileJobRepository) UpdateProfileJobStatus(_ context.Context, id, jobID int, status string) (models.ProfileJob, error) {
	repo.called, repo.id, repo.jobID = true, id, jobID
	repo.status = &status
	return repo.result, repo.err
}

func (repo *stubProfileJobRepository) UpdateProfileJobNotes(_ context.Context, id, jobID int, notes string) (models.ProfileJob, error) {
	repo.called, repo.id, repo.jobID = true, id, jobID
	repo.notes = &notes
	return repo.result, repo.err
}

func (repo *stubProfileJobRepository) GetProfileJobs(_ context.Context, id int) ([]models.SavedProfileJob, error) {
	repo.called = true
	repo.id = id
	return repo.jobs, repo.err
}

func TestSearchJobsHandlerForwardsAndValidatesResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	python := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/search" {
			t.Fatalf("unexpected downstream request: %s %s", r.Method, r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != `{"query":"C++","limit":50}` {
			t.Fatalf("unexpected downstream body: %s", body)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected downstream content type: %s", r.Header.Get("Content-Type"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"jobs":[{"source":"jobtech","external_id":"job-1","title":"C++ Developer","company":"Example AB","location":null,"url":"https://example.com/jobs/1","description":"Build embedded software.","skills":["c++","git"],"date_posted":null,"is_remote":false}]}`))
	}))
	defer python.Close()

	router := gin.New()
	router.POST("/jobs/search", SearchJobsHandler(python.Client(), python.URL))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/jobs/search",
		strings.NewReader(`{"query":"C++","limit":50}`),
	)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	want := `{"count":1,"jobs":[{"source":"jobtech","external_id":"job-1","title":"C++ Developer","company":"Example AB","location":null,"url":"https://example.com/jobs/1","description":"Build embedded software.","skills":["c++","git"],"date_posted":null,"is_remote":false}]}`
	if response.Body.String() != want {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}

func TestSearchJobsHandlerPassesIngestionErrorThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)

	python := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"detail":"invalid request"}`))
	}))
	defer python.Close()

	router := gin.New()
	router.POST("/jobs/search", SearchJobsHandler(python.Client(), python.URL))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/jobs/search", strings.NewReader(`{}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	if response.Body.String() != `{"detail":"invalid request"}` {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}

func TestSearchJobsHandlerRejectsInvalidIngestionResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	python := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer python.Close()

	router := gin.New()
	router.POST("/jobs/search", SearchJobsHandler(python.Client(), python.URL))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/jobs/search", strings.NewReader(`{}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	if response.Body.String() != `{"message":"Job search service returned an invalid response"}` {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}

func TestSearchJobsHandlerReturnsBadGatewayWhenIngestionIsUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	python := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	client := python.Client()
	python.Close()

	router := gin.New()
	router.POST("/jobs/search", SearchJobsHandler(client, python.URL))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/jobs/search", strings.NewReader(`{}`))

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d", response.Code)
	}
	if response.Body.String() != `{"message":"Job search service unavailable"}` {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}

func TestGetProfileJobsHandlerReturnsSavedJobs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	savedAt := time.Date(2026, time.September, 27, 11, 0, 0, 0, time.UTC)
	externalID := "jobtech-123"
	notes := "Need to work on SQL joins"
	repo := &stubProfileJobRepository{jobs: []models.SavedProfileJob{{
		JobID: 7,
		Job: models.JobSearchResult{
			Source:     "jobtech",
			ExternalID: &externalID,
			Title:      "Backend Developer",
			Skills:     []string{"go", "postgresql"},
		},
		Status:    "applied",
		Notes:     &notes,
		SavedAt:   savedAt,
		UpdatedAt: savedAt,
	}}}
	router := gin.New()
	router.GET("/profiles/:id/jobs", GetProfileJobsHandler(repo))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/profiles/42/jobs", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d; body: %s", response.Code, response.Body.String())
	}
	var body struct {
		ProfileJob []models.SavedProfileJob `json:"profile_job"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.ProfileJob) != 1 {
		t.Fatalf("expected one saved job, got %+v", body.ProfileJob)
	}
	job := body.ProfileJob[0]
	if job.JobID != 7 || job.Job.Title != "Backend Developer" || job.Job.ExternalID == nil || *job.Job.ExternalID != externalID || job.Status != "applied" || job.Notes == nil || *job.Notes != notes || !job.SavedAt.Equal(savedAt) || !job.UpdatedAt.Equal(savedAt) {
		t.Errorf("unexpected saved job: %+v", job)
	}
	if len(job.Job.Skills) != 2 || job.Job.Skills[0] != "go" || job.Job.Skills[1] != "postgresql" {
		t.Errorf("unexpected skills: %v", job.Job.Skills)
	}
	if !repo.called || repo.id != 42 {
		t.Errorf("unexpected repository call: called=%v id=%d", repo.called, repo.id)
	}
}

func TestGetProfileJobsHandlerReturnsEmptyList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubProfileJobRepository{jobs: []models.SavedProfileJob{}}
	router := gin.New()
	router.GET("/profiles/:id/jobs", GetProfileJobsHandler(repo))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/profiles/42/jobs", nil))

	assertJSONResponse(t, response, http.StatusOK, `{"profile_job":[]}`)
	if !repo.called || repo.id != 42 {
		t.Errorf("unexpected repository call: called=%v id=%d", repo.called, repo.id)
	}
}

func TestGetProfileJobsHandlerRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, id := range []string{"invalid", "0", "-1"} {
		t.Run(id, func(t *testing.T) {
			repo := &stubProfileJobRepository{}
			router := gin.New()
			router.GET("/profiles/:id/jobs", GetProfileJobsHandler(repo))
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/profiles/"+id+"/jobs", nil))

			assertJSONResponse(t, response, http.StatusBadRequest, `{"message":"A valid profileID is required"}`)
			if repo.called {
				t.Fatal("expected GetProfileJobs not to be called")
			}
		})
	}
}

func TestGetProfileJobsHandlerReturnsInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubProfileJobRepository{err: errors.New("database unavailable")}
	router := gin.New()
	router.GET("/profiles/:id/jobs", GetProfileJobsHandler(repo))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/profiles/42/jobs", nil))

	assertJSONResponse(t, response, http.StatusInternalServerError, `{"message":"Failed to get profile jobs"}`)
	if !repo.called || repo.id != 42 {
		t.Errorf("unexpected repository call: called=%v id=%d", repo.called, repo.id)
	}
}

func TestUpdateProfileJobHandlerStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubProfileJobRepository{result: models.ProfileJob{ProfileID: 42, JobID: 7, Status: "applied"}}
	router := gin.New()
	router.PUT("/profiles/:id/jobs/:jobId", UpdateProfileJobHandler(repo))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/profiles/42/jobs/7", strings.NewReader(`{"status":"applied"}`)))

	assertJSONResponse(t, response, http.StatusOK, `{"profile_job":{"profile_id":42,"job_id":7,"status":"applied","notes":null,"saved_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}}`)
	if !repo.called || repo.id != 42 || repo.jobID != 7 || repo.status == nil || *repo.status != "applied" || repo.notes != nil {
		t.Errorf("unexpected repository call: called=%v id=%d", repo.called, repo.id)
	}
}

func TestUpdateProfileJobHandlerStatusAndNotes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	notes := "Review SQL joins"
	repo := &stubProfileJobRepository{result: models.ProfileJob{ProfileID: 42, JobID: 7, Status: "interviewing", Notes: &notes}}
	router := gin.New()
	router.PUT("/profiles/:id/jobs/:jobId", UpdateProfileJobHandler(repo))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/profiles/42/jobs/7", strings.NewReader(`{"status":"interviewing","notes":"Review SQL joins"}`)))

	assertJSONResponse(t, response, http.StatusOK, `{"profile_job":{"profile_id":42,"job_id":7,"status":"interviewing","notes":"Review SQL joins","saved_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}}`)
	if !repo.called || repo.id != 42 || repo.jobID != 7 || repo.notes == nil || *repo.notes != notes || repo.status == nil || *repo.status != "interviewing" {
		t.Errorf("unexpected repository call: %+v", repo)
	}
}

func TestUpdateProfileJobHandlerRejectsInvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubProfileJobRepository{}
	router := gin.New()
	router.PUT("/profiles/:id/jobs/:jobId", UpdateProfileJobHandler(repo))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/profiles/42/jobs/7", strings.NewReader(`{"status":"invalid"}`)))

	assertJSONResponse(t, response, http.StatusBadRequest, `{"message":"A valid job status is required"}`)
	if repo.called {
		t.Fatal("expected UpdateProfileJobStatus not to be called")
	}
}

func TestUpdateProfileJobHandlerRejectsMissingStatusAndNotes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubProfileJobRepository{}
	router := gin.New()
	router.PUT("/profiles/:id/jobs/:jobId", UpdateProfileJobHandler(repo))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/profiles/42/jobs/7", strings.NewReader(`{}`)))

	assertJSONResponse(t, response, http.StatusBadRequest, `{"message":"At least one of status or notes is required"}`)
	if repo.called {
		t.Fatal("expected UpdateProfileJobStatus not to be called")
	}
}
