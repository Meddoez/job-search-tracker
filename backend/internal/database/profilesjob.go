package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Meddoez/job-search-tracker/backend/internal/models"
	"github.com/lib/pq"
)

// PostgresProfileJobRepository stores and retrieves profile-job associations in PostgreSQL.
type PostgresProfileJobRepository struct {
	db *sql.DB
}

// NewProfileJobRepository returns a PostgreSQL-backed profile-job repository.
func NewProfileJobRepository(db *sql.DB) *PostgresProfileJobRepository {
	return &PostgresProfileJobRepository{db: db}
}

// SaveProfileJob stores a normalized job, its skills, and its profile association atomically.
func (repo *PostgresProfileJobRepository) SaveProfileJob(ctx context.Context, profileID int, job models.JobSearchResult) (models.ProfileJob, error) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return models.ProfileJob{}, fmt.Errorf("begin save profile job: %w", err)
	}
	defer tx.Rollback()

	const upsertJobQuery = `
		INSERT INTO jobs (
			source,
			external_id,
			title,
			company,
			location,
			url,
			description,
			date_posted,
			is_remote
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (source, external_id) WHERE external_id IS NOT NULL
		DO UPDATE SET
			title = EXCLUDED.title,
			company = EXCLUDED.company,
			location = EXCLUDED.location,
			url = EXCLUDED.url,
			description = EXCLUDED.description,
			date_posted = EXCLUDED.date_posted,
			is_remote = EXCLUDED.is_remote
		RETURNING id
	`

	var jobID int
	err = tx.QueryRowContext(
		ctx,
		upsertJobQuery,
		job.Source,
		job.ExternalID,
		job.Title,
		job.Company,
		job.Location,
		job.URL,
		job.Description,
		job.DatePosted,
		job.IsRemote,
	).Scan(&jobID)
	if err != nil {
		return models.ProfileJob{}, fmt.Errorf("upsert job: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM job_skills WHERE job_id = $1", jobID); err != nil {
		return models.ProfileJob{}, fmt.Errorf("delete job skills: %w", err)
	}

	for position, skill := range job.Skills {
		_, err := tx.ExecContext(
			ctx,
			"INSERT INTO job_skills (job_id, skill, position) VALUES ($1, $2, $3)",
			jobID,
			skill,
			position,
		)
		if err != nil {
			return models.ProfileJob{}, fmt.Errorf("insert job skill: %w", err)
		}
	}

	const saveProfileJobQuery = `
		INSERT INTO profile_jobs (profile_id, job_id)
		VALUES ($1, $2)
		ON CONFLICT (profile_id, job_id)
		DO UPDATE SET job_id = profile_jobs.job_id
		RETURNING profile_id, job_id, status, saved_at, updated_at
	`

	var profileJob models.ProfileJob
	err = tx.QueryRowContext(ctx, saveProfileJobQuery, profileID, jobID).Scan(
		&profileJob.ProfileID,
		&profileJob.JobID,
		&profileJob.Status,
		&profileJob.SavedAt,
		&profileJob.UpdatedAt,
	)
	if err != nil {
		return models.ProfileJob{}, fmt.Errorf("insert profile job: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return models.ProfileJob{}, fmt.Errorf("commit profile job: %w", err)
	}

	return profileJob, nil
}

func (repo *PostgresProfileJobRepository) UpdateProfileJobStatus(ctx context.Context, profileID int, jobID int, status string) (models.ProfileJob, error) {
	query := `
	UPDATE profile_jobs
	SET
    status = $3,
    updated_at = NOW()
	WHERE profile_id = $1
	AND job_id = $2
	RETURNING profile_id, job_id, status, saved_at, updated_at
	`

	var profileJob models.ProfileJob
	err := repo.db.QueryRowContext(ctx, query, profileID, jobID, status).Scan(
		&profileJob.ProfileID,
		&profileJob.JobID,
		&profileJob.Status,
		&profileJob.SavedAt,
		&profileJob.UpdatedAt,
	)
	if err != nil {
		return models.ProfileJob{}, fmt.Errorf("Update profile job status: %w", err)
	}

	return profileJob, nil
}

func (repo *PostgresProfileJobRepository) GetProfileJobs(ctx context.Context, profileID int) ([]models.SavedProfileJob, error) {
	query := `
		SELECT
		j.id,
		j.source,
		j.external_id,
		j.title,
		j.company,
		j.location,
		j.url,
		j.description,
		j.date_posted,
		j.is_remote,
		ARRAY(
			SELECT js.skill
			FROM job_skills AS js
			WHERE js.job_id = j.id
			ORDER BY js.position
		) AS skills,
		pj.status,
		pj.saved_at,
		pj.updated_at
	FROM profile_jobs AS pj
	JOIN jobs AS j ON j.id = pj.job_id
	WHERE pj.profile_id = $1
	ORDER BY pj.saved_at DESC, j.id DESC;`

	rows, err := repo.db.QueryContext(ctx, query, profileID)
	if err != nil {
		return []models.SavedProfileJob{}, err
	}
	defer rows.Close()

	savedProfileJobs := make([]models.SavedProfileJob, 0)
	for rows.Next() {
		var saved models.SavedProfileJob
		err := rows.Scan(
			&saved.JobID,
			&saved.Job.Source,
			&saved.Job.ExternalID,
			&saved.Job.Title,
			&saved.Job.Company,
			&saved.Job.Location,
			&saved.Job.URL,
			&saved.Job.Description,
			&saved.Job.DatePosted,
			&saved.Job.IsRemote,
			pq.Array(&saved.Job.Skills),
			&saved.Status,
			&saved.SavedAt,
			&saved.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		savedProfileJobs = append(savedProfileJobs, saved)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return savedProfileJobs, nil
}
