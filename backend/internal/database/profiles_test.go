package database

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

const (
	createProfileQuery = "INSERT INTO profiles(name) VALUES($1) RETURNING id, name, created_at, updated_at"
	addSkillQuery      = "INSERT INTO profile_skills(profile_id, skill) VALUES($1, $2) ON CONFLICT(profile_id, skill) DO NOTHING"
	getProfileQuery    = `
		SELECT
			p.id,
			p.name,
			ARRAY(
				SELECT ps.skill
				FROM profile_skills AS ps
				WHERE ps.profile_id = p.id
				ORDER BY ps.skill
			),
			p.created_at,
			p.updated_at
		FROM profiles AS p
		WHERE p.id = $1
	`
	getProfilesQuery = `
		SELECT
			p.id,
			p.name,
			ARRAY(
				SELECT ps.skill
				FROM profile_skills AS ps
				WHERE ps.profile_id = p.id
				ORDER BY ps.skill
			),
			p.created_at,
			p.updated_at
		FROM profiles AS p
		ORDER BY p.id DESC
	`
)

func TestCreateProfileReturnsCreatedProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewProfileRepository(db)

	createdAt := time.Date(2026, time.September, 10, 8, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	skills := []string{"go", "postgresql"}
	rows := sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
		AddRow(7, "Backend", createdAt, updatedAt)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(createProfileQuery)).
		WithArgs("Backend").
		WillReturnRows(rows)
	mock.ExpectExec(regexp.QuoteMeta(addSkillQuery)).
		WithArgs(7, "go").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(addSkillQuery)).
		WithArgs(7, "postgresql").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	profile, err := repo.CreateProfile(context.Background(), "Backend", skills)
	if err != nil {
		t.Fatalf("CreateProfile returned an error: %v", err)
	}
	if profile.ID != 7 || profile.Name != "Backend" || !profile.CreatedAt.Equal(createdAt) || !profile.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	if len(profile.Skills) != 2 || profile.Skills[0] != "go" || profile.Skills[1] != "postgresql" {
		t.Fatalf("unexpected profile skills: %#v", profile.Skills)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateProfileReturnsDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewProfileRepository(db)

	wantErr := errors.New("insert failed")
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(createProfileQuery)).
		WithArgs("Backend").
		WillReturnError(wantErr)
	mock.ExpectRollback()

	profile, err := repo.CreateProfile(context.Background(), "Backend", nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if profile.ID != 0 {
		t.Fatalf("expected a zero profile, got %+v", profile)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetProfileReturnsMatchingProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewProfileRepository(db)

	createdAt := time.Date(2026, time.September, 10, 8, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	rows := sqlmock.NewRows([]string{"id", "name", "skills", "created_at", "updated_at"}).
		AddRow(7, "Backend", "{go,postgresql}", createdAt, updatedAt)
	mock.ExpectQuery(regexp.QuoteMeta(getProfileQuery)).
		WithArgs(7).
		WillReturnRows(rows)

	profile, err := repo.GetProfile(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetProfile returned an error: %v", err)
	}
	if profile.ID != 7 || profile.Name != "Backend" || !profile.CreatedAt.Equal(createdAt) || !profile.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	if len(profile.Skills) != 2 || profile.Skills[0] != "go" || profile.Skills[1] != "postgresql" {
		t.Fatalf("unexpected profile skills: %#v", profile.Skills)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetProfileReturnsEmptySkills(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewProfileRepository(db)

	createdAt := time.Date(2026, time.September, 10, 8, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	rows := sqlmock.NewRows([]string{"id", "name", "skills", "created_at", "updated_at"}).
		AddRow(7, "Backend", "{}", createdAt, updatedAt)
	mock.ExpectQuery(regexp.QuoteMeta(getProfileQuery)).
		WithArgs(7).
		WillReturnRows(rows)

	profile, err := repo.GetProfile(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetProfile returned an error: %v", err)
	}
	if profile.Skills == nil || len(profile.Skills) != 0 {
		t.Fatalf("expected non-nil empty skills, got %#v", profile.Skills)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetProfileReturnsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewProfileRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(getProfileQuery)).
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	profile, err := repo.GetProfile(context.Background(), 99)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
	if profile.ID != 0 {
		t.Fatalf("expected a zero profile, got %+v", profile)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetProfilesReturnsAllProfiles(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewProfileRepository(db)

	createdAt := time.Date(2026, time.September, 10, 8, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "name", "skills", "created_at", "updated_at"}).
		AddRow(8, "Frontend", "{react,typescript}", createdAt.Add(time.Hour), createdAt.Add(time.Hour)).
		AddRow(7, "Backend", "{go,postgresql}", createdAt, createdAt)
	mock.ExpectQuery(regexp.QuoteMeta(getProfilesQuery)).WillReturnRows(rows)

	profiles, err := repo.GetProfiles(context.Background())
	if err != nil {
		t.Fatalf("GetProfiles returned an error: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
	if profiles[0].ID != 8 || profiles[0].Name != "Frontend" || len(profiles[0].Skills) != 2 || profiles[0].Skills[0] != "react" || profiles[0].Skills[1] != "typescript" {
		t.Fatalf("unexpected first profile: %+v", profiles[0])
	}
	if profiles[1].ID != 7 || profiles[1].Name != "Backend" || len(profiles[1].Skills) != 2 || profiles[1].Skills[0] != "go" || profiles[1].Skills[1] != "postgresql" {
		t.Fatalf("unexpected second profile: %+v", profiles[1])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetProfilesReturnsEmptySlice(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewProfileRepository(db)

	rows := sqlmock.NewRows([]string{"id", "name", "skills", "created_at", "updated_at"})
	mock.ExpectQuery(regexp.QuoteMeta(getProfilesQuery)).WillReturnRows(rows)

	profiles, err := repo.GetProfiles(context.Background())
	if err != nil {
		t.Fatalf("GetProfiles returned an error: %v", err)
	}
	if profiles == nil || len(profiles) != 0 {
		t.Fatalf("expected a non-nil empty slice, got %#v", profiles)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetProfilesReturnsDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewProfileRepository(db)

	wantErr := errors.New("select failed")
	mock.ExpectQuery(regexp.QuoteMeta(getProfilesQuery)).WillReturnError(wantErr)

	_, err = repo.GetProfiles(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteProfileReturnsNoError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewProfileRepository(db)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM profiles WHERE id = $1")).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.DeleteProfile(context.Background(), 7)
	if err != nil {
		t.Fatalf("DeleteProfile returned an error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
