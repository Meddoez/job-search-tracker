package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Meddoez/job-search-tracker/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type stubProfileRepository struct {
	profile      models.Profile
	profiles     []models.Profile
	err          error
	createCalled bool
	createName   string
	createSkills []string
	getCalled    bool
	getID        int
	getAllCalled bool
}

func (repo *stubProfileRepository) CreateProfile(_ context.Context, name string, skills []string) (models.Profile, error) {
	repo.createCalled = true
	repo.createName = name
	repo.createSkills = skills
	return repo.profile, repo.err
}

func (repo *stubProfileRepository) GetProfile(_ context.Context, id int) (models.Profile, error) {
	repo.getCalled = true
	repo.getID = id
	return repo.profile, repo.err
}

func (repo *stubProfileRepository) GetProfiles(context.Context) ([]models.Profile, error) {
	repo.getAllCalled = true
	return repo.profiles, repo.err
}

func (repo *stubProfileRepository) DeleteProfile(_ context.Context, id int) error {
	repo.getCalled = true
	repo.getID = id
	return repo.err
}

func assertJSONResponse(t *testing.T, response *httptest.ResponseRecorder, status int, body string) {
	t.Helper()
	if response.Code != status {
		t.Fatalf("unexpected status: %d; body: %s", response.Code, response.Body.String())
	}
	if response.Body.String() != body {
		t.Fatalf("unexpected response body: %s", response.Body.String())
	}
}

func TestCreateProfileHandlerCreatesProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	createdAt := time.Date(2026, time.September, 10, 8, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	repo := &stubProfileRepository{profile: models.Profile{
		ID:        7,
		Name:      "Backend",
		Skills:    []string{"go", "postgresql"},
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}}

	router := gin.New()
	router.POST("/profiles", CreateProfileHandler(repo))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/profiles", strings.NewReader(`{"name":"Backend","skills":["go","postgresql"]}`))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d; body: %s", response.Code, response.Body.String())
	}
	var body struct {
		Profile models.Profile `json:"profile"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Profile.ID != 7 || body.Profile.Name != "Backend" || !body.Profile.CreatedAt.Equal(createdAt) || !body.Profile.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("unexpected profile: %+v", body.Profile)
	}
	if len(body.Profile.Skills) != 2 || body.Profile.Skills[0] != "go" || body.Profile.Skills[1] != "postgresql" {
		t.Fatalf("unexpected profile skills: %#v", body.Profile.Skills)
	}
	if !repo.createCalled || repo.createName != "Backend" {
		t.Fatalf("unexpected repository call: called=%v name=%q", repo.createCalled, repo.createName)
	}
	if len(repo.createSkills) != 2 || repo.createSkills[0] != "go" || repo.createSkills[1] != "postgresql" {
		t.Fatalf("unexpected repository skills: %#v", repo.createSkills)
	}
}

func TestCreateProfileHandlerRejectsInvalidName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := map[string]string{
		"malformed JSON": `{`,
		"missing name":   `{}`,
		"empty name":     `{"name":""}`,
		"blank name":     `{"name":"   "}`,
	}

	for name, requestBody := range tests {
		t.Run(name, func(t *testing.T) {
			repo := &stubProfileRepository{}
			router := gin.New()
			router.POST("/profiles", CreateProfileHandler(repo))
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/profiles", strings.NewReader(requestBody))
			request.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(response, request)

			assertJSONResponse(t, response, http.StatusBadRequest, `{"message":"A profile name is required"}`)
			if repo.createCalled {
				t.Fatal("expected CreateProfile not to be called")
			}
		})
	}
}

func TestCreateProfileHandlerReturnsInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubProfileRepository{err: errors.New("insert failed")}

	router := gin.New()
	router.POST("/profiles", CreateProfileHandler(repo))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/profiles", strings.NewReader(`{"name":"Backend"}`))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(response, request)

	assertJSONResponse(t, response, http.StatusInternalServerError, `{"message":"Failed to create profile"}`)
	if !repo.createCalled || repo.createName != "Backend" {
		t.Fatalf("unexpected repository call: called=%v name=%q", repo.createCalled, repo.createName)
	}
}

func TestCreateProfileHandlerRejectsDuplicateName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubProfileRepository{err: &pq.Error{
		Code:       "23505",
		Constraint: "uq_profiles_name_normalized",
	}}

	router := gin.New()
	router.POST("/profiles", CreateProfileHandler(repo))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/profiles", strings.NewReader(`{"name":"backend"}`))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(response, request)

	assertJSONResponse(t, response, http.StatusConflict, `{"message":"A profile with this name already exists"}`)
	if !repo.createCalled || repo.createName != "backend" {
		t.Fatalf("unexpected repository call: called=%v name=%q", repo.createCalled, repo.createName)
	}
}

func TestGetProfileHandlerReturnsProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	createdAt := time.Date(2026, time.September, 10, 8, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	repo := &stubProfileRepository{profile: models.Profile{
		ID:        7,
		Name:      "Backend",
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}}

	router := gin.New()
	router.GET("/profiles/:id", GetProfileHandler(repo))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/profiles/7", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d; body: %s", response.Code, response.Body.String())
	}
	var body struct {
		Profile models.Profile `json:"profile"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Profile.ID != 7 || body.Profile.Name != "Backend" {
		t.Fatalf("unexpected profile: %+v", body.Profile)
	}
	if !repo.getCalled || repo.getID != 7 {
		t.Fatalf("unexpected repository call: called=%v id=%d", repo.getCalled, repo.getID)
	}
}

func TestGetProfileHandlerRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, id := range []string{"invalid", "0", "-1"} {
		t.Run(id, func(t *testing.T) {
			repo := &stubProfileRepository{}
			router := gin.New()
			router.GET("/profiles/:id", GetProfileHandler(repo))
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/profiles/"+id, nil)

			router.ServeHTTP(response, request)

			assertJSONResponse(t, response, http.StatusBadRequest, `{"message":"A valid profile ID is required"}`)
			if repo.getCalled {
				t.Fatal("expected GetProfile not to be called")
			}
		})
	}
}

func TestGetProfileHandlerReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubProfileRepository{err: sql.ErrNoRows}

	router := gin.New()
	router.GET("/profiles/:id", GetProfileHandler(repo))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/profiles/99", nil)

	router.ServeHTTP(response, request)

	assertJSONResponse(t, response, http.StatusNotFound, `{"message":"Profile not found"}`)
	if !repo.getCalled || repo.getID != 99 {
		t.Fatalf("unexpected repository call: called=%v id=%d", repo.getCalled, repo.getID)
	}
}

func TestGetProfileHandlerReturnsInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubProfileRepository{err: errors.New("database unavailable")}

	router := gin.New()
	router.GET("/profiles/:id", GetProfileHandler(repo))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/profiles/7", nil)

	router.ServeHTTP(response, request)

	assertJSONResponse(t, response, http.StatusInternalServerError, `{"message":"Failed to get profile"}`)
	if !repo.getCalled || repo.getID != 7 {
		t.Fatalf("unexpected repository call: called=%v id=%d", repo.getCalled, repo.getID)
	}
}

func TestGetProfilesHandlerReturnsProfiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	createdAt := time.Date(2026, time.September, 10, 8, 0, 0, 0, time.UTC)
	repo := &stubProfileRepository{profiles: []models.Profile{
		{ID: 8, Name: "Frontend", Skills: []string{"react", "typescript"}, CreatedAt: createdAt, UpdatedAt: createdAt},
		{ID: 7, Name: "Backend", Skills: []string{"go", "postgresql"}, CreatedAt: createdAt, UpdatedAt: createdAt},
	}}

	router := gin.New()
	router.GET("/profiles", GetProfilesHandler(repo))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/profiles", nil)

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d; body: %s", response.Code, response.Body.String())
	}
	var body struct {
		Profiles []models.Profile `json:"profiles"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Profiles) != 2 || body.Profiles[0].ID != 8 || body.Profiles[1].ID != 7 {
		t.Fatalf("unexpected profiles: %+v", body.Profiles)
	}
	if len(body.Profiles[0].Skills) != 2 || body.Profiles[0].Skills[0] != "react" || body.Profiles[0].Skills[1] != "typescript" {
		t.Fatalf("unexpected profile skills: %#v", body.Profiles[0].Skills)
	}
	if !repo.getAllCalled {
		t.Fatal("expected GetProfiles to be called")
	}
}

func TestGetProfilesHandlerReturnsEmptyList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubProfileRepository{profiles: []models.Profile{}}

	router := gin.New()
	router.GET("/profiles", GetProfilesHandler(repo))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/profiles", nil)

	router.ServeHTTP(response, request)

	assertJSONResponse(t, response, http.StatusOK, `{"profiles":[]}`)
	if !repo.getAllCalled {
		t.Fatal("expected GetProfiles to be called")
	}
}

func TestGetProfilesHandlerReturnsInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &stubProfileRepository{err: errors.New("database unavailable")}

	router := gin.New()
	router.GET("/profiles", GetProfilesHandler(repo))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/profiles", nil)

	router.ServeHTTP(response, request)

	assertJSONResponse(t, response, http.StatusInternalServerError, `{"message":"Failed to get profiles"}`)
	if !repo.getAllCalled {
		t.Fatal("expected GetProfiles to be called")
	}
}
