package database

import (
	"context"
	"database/sql"

	"github.com/Meddoez/job-search-tracker/backend/internal/models"
	"github.com/lib/pq"
)

// PostgresProfileRepository stores and retrieves skill profiles in PostgreSQL.
type PostgresProfileRepository struct {
	db *sql.DB
}

// NewProfileRepository returns a PostgreSQL-backed profile repository.
func NewProfileRepository(db *sql.DB) *PostgresProfileRepository {
	return &PostgresProfileRepository{db: db}
}

// CreateProfile creates a profile and its skills in one transaction.
func (repo *PostgresProfileRepository) CreateProfile(ctx context.Context, name string, skills []string) (models.Profile, error) {
	rows := "INSERT INTO profiles(name) VALUES($1) RETURNING id, name, created_at, updated_at"

	tx, err := repo.db.BeginTx(ctx, nil)

	if err != nil {
		return models.Profile{}, err
	}
	defer tx.Rollback()

	var profile models.Profile
	err = tx.QueryRowContext(ctx, rows, name).Scan(
		&profile.ID,
		&profile.Name,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		return models.Profile{}, err
	}

	err = repo.AddSkillsById(ctx, tx, profile.ID, skills)
	if err != nil {
		return models.Profile{}, err
	}

	profile.Skills = skills

	err = tx.Commit()
	if err != nil {
		return models.Profile{}, err
	}

	return profile, nil
}

// GetProfile returns the profile with the given ID.
// It returns sql.ErrNoRows when the profile does not exist.
func (repo *PostgresProfileRepository) GetProfile(ctx context.Context, id int) (models.Profile, error) {
	rows := `
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

	return scanProfile(repo.db.QueryRowContext(ctx, rows, id))
}

// GetProfiles returns all profiles ordered from newest to oldest.
// Skills within each profile are ordered alphabetically.
func (repo *PostgresProfileRepository) GetProfiles(ctx context.Context) ([]models.Profile, error) {
	query := `
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

	rows, err := repo.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make([]models.Profile, 0)
	for rows.Next() {
		profile, err := scanProfile(rows)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return profiles, nil
}

type profileScanner interface {
	Scan(dest ...any) error
}

func scanProfile(scanner profileScanner) (models.Profile, error) {
	var profile models.Profile
	err := scanner.Scan(
		&profile.ID,
		&profile.Name,
		pq.Array(&profile.Skills),
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		return profile, err
	}

	return profile, nil
}

// AddSkillsById adds skills to a profile within tx and ignores existing profile-skill pairs.
func (repo *PostgresProfileRepository) AddSkillsById(ctx context.Context, tx *sql.Tx, id int, skills []string) error {
	rows := "INSERT INTO profile_skills(profile_id, skill) VALUES($1, $2) ON CONFLICT(profile_id, skill) DO NOTHING"

	for _, skill := range skills {
		_, err := tx.ExecContext(ctx, rows, id, skill)
		if err != nil {
			return err
		}
	}

	return nil
}

func (repo *PostgresProfileRepository) DeleteProfile(ctx context.Context, id int) error {
	query := "DELETE FROM profiles WHERE id = $1"
	result, err := repo.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return sql.ErrNoRows
	}

	return nil
}
