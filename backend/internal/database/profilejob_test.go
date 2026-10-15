package database

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Meddoez/job-search-tracker/backend/internal/models"
)

const (
	profileID = 1
	jobID     = 7
)

func TestSaveProfileJobReturnsSavedProfileJob(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	externalID := "jobtech-123"
	company := "Example AB"
	location := "Stockholm"
	url := "https://example.com/jobs/123"
	description := "Build Go services."
	datePosted := "2026-09-27T10:00:00"
	isRemote := true
	job := models.JobSearchResult{
		Source:      "jobtech",
		ExternalID:  &externalID,
		Title:       "Backend Developer",
		Company:     &company,
		Location:    &location,
		URL:         &url,
		Description: &description,
		Skills:      []string{"go", "postgresql"},
		DatePosted:  &datePosted,
		IsRemote:    &isRemote,
	}
	savedAt := time.Date(2026, time.September, 27, 11, 0, 0, 0, time.UTC)
	updatedAt := savedAt

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO jobs")).
		WithArgs(
			job.Source,
			externalID,
			job.Title,
			company,
			location,
			url,
			description,
			datePosted,
			isRemote,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(jobID))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM job_skills WHERE job_id = $1")).
		WithArgs(jobID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO job_skills (job_id, skill, position) VALUES ($1, $2, $3)")).
		WithArgs(jobID, "go", 0).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO job_skills (job_id, skill, position) VALUES ($1, $2, $3)")).
		WithArgs(jobID, "postgresql", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO profile_jobs")).
		WithArgs(profileID, jobID).
		WillReturnRows(
			sqlmock.NewRows([]string{"profile_id", "job_id", "status", "notes", "saved_at", "updated_at"}).
				AddRow(profileID, jobID, "saved", nil, savedAt, updatedAt),
		)
	mock.ExpectCommit()

	repo := NewProfileJobRepository(db)
	profileJob, err := repo.SaveProfileJob(context.Background(), profileID, job)
	if err != nil {
		t.Fatal(err)
	}
	if profileJob.ProfileID != profileID || profileJob.JobID != jobID || profileJob.Status != "saved" {
		t.Errorf("unexpected profile job: %+v", profileJob)
	}
	if !profileJob.SavedAt.Equal(savedAt) || !profileJob.UpdatedAt.Equal(updatedAt) {
		t.Errorf("unexpected profile job timestamps: %+v", profileJob)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateProfileJobStatusReturnsUpdatedProfileJob(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	savedAt := time.Date(2026, time.September, 27, 11, 0, 0, 0, time.UTC)
	updatedAt := savedAt.Add(time.Hour)
	newStatus := "applied"
	notes := "Review SQL joins"
	mock.ExpectQuery(regexp.QuoteMeta("UPDATE profile_jobs SET status = $3, updated_at = NOW() WHERE profile_id = $1 AND job_id = $2")).
		WithArgs(profileID, jobID, newStatus).
		WillReturnRows(sqlmock.NewRows([]string{"profile_id", "job_id", "status", "notes", "saved_at", "updated_at"}).
			AddRow(profileID, jobID, newStatus, notes, savedAt, updatedAt))

	repo := NewProfileJobRepository(db)
	profileJob, err := repo.UpdateProfileJobStatus(context.Background(), profileID, jobID, newStatus)
	if err != nil {
		t.Fatal(err)
	}
	if profileJob.ProfileID != profileID || profileJob.JobID != jobID || profileJob.Status != newStatus {
		t.Errorf("unexpected profile job: %+v", profileJob)
	}
	if profileJob.Notes == nil || *profileJob.Notes != notes {
		t.Errorf("expected existing notes to be returned, got %+v", profileJob)
	}
	if !profileJob.SavedAt.Equal(savedAt) || !profileJob.UpdatedAt.Equal(updatedAt) {
		t.Errorf("unexpected profile job timestamps: %+v", profileJob)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateProfileJobNotes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	notes := "Review SQL joins"
	savedAt := time.Date(2026, time.September, 27, 11, 0, 0, 0, time.UTC)
	updatedAt := savedAt.Add(time.Hour)
	mock.ExpectQuery(regexp.QuoteMeta("UPDATE profile_jobs SET notes = $3, updated_at = NOW() WHERE profile_id = $1 AND job_id = $2")).
		WithArgs(profileID, jobID, notes).
		WillReturnRows(sqlmock.NewRows([]string{"profile_id", "job_id", "status", "notes", "saved_at", "updated_at"}).
			AddRow(profileID, jobID, "interviewing", notes, savedAt, updatedAt))

	job, err := NewProfileJobRepository(db).UpdateProfileJobNotes(context.Background(), profileID, jobID, notes)
	if err != nil {
		t.Fatal(err)
	}
	if job.ProfileID != profileID || job.JobID != jobID || job.Status != "interviewing" || job.Notes == nil || *job.Notes != notes || !job.SavedAt.Equal(savedAt) || !job.UpdatedAt.Equal(updatedAt) {
		t.Errorf("unexpected updated profile job: %+v", job)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetProfileJobsReturnsSavedJobs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	savedAt := time.Date(2026, time.September, 27, 11, 0, 0, 0, time.UTC)
	updatedAt := savedAt.Add(time.Hour)
	mock.ExpectQuery(regexp.QuoteMeta("FROM profile_jobs AS pj")).
		WithArgs(profileID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "source", "external_id", "title", "company", "location", "url",
			"description", "date_posted", "is_remote", "skills", "status", "notes", "saved_at", "updated_at",
		}).AddRow(
			jobID, "jobtech", "jobtech-123", "Backend Developer", "Example AB", "Stockholm",
			"https://example.com/jobs/123", "Build Go services.", "2026-09-27T10:00:00", true,
			"{go,postgresql}", "applied", "Need to work on SQL joins", savedAt, updatedAt,
		).AddRow(
			8, "jobtech", "jobtech-456", "Frontend Developer", nil, nil,
			nil, nil, nil, nil, "{}", "saved", "Need to review frontend frameworks", savedAt, savedAt,
		))

	jobs, err := NewProfileJobRepository(db).GetProfileJobs(context.Background(), profileID)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected two saved jobs, got %+v", jobs)
	}
	first := jobs[0]
	if first.JobID != jobID || first.Job.Source != "jobtech" || first.Job.ExternalID == nil || *first.Job.ExternalID != "jobtech-123" || first.Job.Title != "Backend Developer" {
		t.Errorf("unexpected first job identity: %+v", first)
	}
	if first.Job.Company == nil || *first.Job.Company != "Example AB" || first.Job.Location == nil || *first.Job.Location != "Stockholm" || first.Job.URL == nil || *first.Job.URL != "https://example.com/jobs/123" || first.Job.Description == nil || *first.Job.Description != "Build Go services." || first.Job.DatePosted == nil || *first.Job.DatePosted != "2026-09-27T10:00:00" || first.Job.IsRemote == nil || !*first.Job.IsRemote {
		t.Errorf("unexpected first job details: %+v", first.Job)
	}
	if len(first.Job.Skills) != 2 || first.Job.Skills[0] != "go" || first.Job.Skills[1] != "postgresql" {
		t.Errorf("unexpected skills: %v", first.Job.Skills)
	}
	if first.Status != "applied" || !first.SavedAt.Equal(savedAt) || !first.UpdatedAt.Equal(updatedAt) {
		t.Errorf("unexpected saved job metadata: %+v", first)
	}
	second := jobs[1]
	if second.JobID != 8 || second.Job.Title != "Frontend Developer" || second.Job.Company != nil || second.Job.IsRemote != nil || len(second.Job.Skills) != 0 || second.Status != "saved" {
		t.Errorf("unexpected second job: %+v", second)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetProfileJobsReturnsEmptyList(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectQuery(regexp.QuoteMeta("FROM profile_jobs AS pj")).
		WithArgs(profileID).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "source", "external_id", "title", "company", "location", "url",
			"description", "date_posted", "is_remote", "skills", "status", "saved_at", "updated_at",
		}))

	jobs, err := NewProfileJobRepository(db).GetProfileJobs(context.Background(), profileID)
	if err != nil {
		t.Fatal(err)
	}
	if jobs == nil || len(jobs) != 0 {
		t.Fatalf("expected an empty non-nil list, got %+v", jobs)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetProfileJobsReturnsQueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	wantErr := errors.New("database unavailable")
	mock.ExpectQuery(regexp.QuoteMeta("FROM profile_jobs AS pj")).
		WithArgs(profileID).
		WillReturnError(wantErr)

	jobs, err := NewProfileJobRepository(db).GetProfileJobs(context.Background(), profileID)
	if !errors.Is(err, wantErr) || len(jobs) != 0 {
		t.Fatalf("expected query error and no jobs, got jobs=%+v err=%v", jobs, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetProfileJobsReturnsScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectQuery(regexp.QuoteMeta("FROM profile_jobs AS pj")).
		WithArgs(profileID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(jobID))

	jobs, err := NewProfileJobRepository(db).GetProfileJobs(context.Background(), profileID)
	if err == nil || len(jobs) != 0 {
		t.Fatalf("expected scan error and no jobs, got jobs=%+v err=%v", jobs, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
