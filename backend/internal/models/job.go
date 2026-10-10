// Package models defines data shared across backend layers.
package models

import "time"

// JobSearchResult is a transient normalized posting returned by ingestion.
type JobSearchResult struct {
	Source      string   `json:"source"`
	ExternalID  *string  `json:"external_id"`
	Title       string   `json:"title"`
	Company     *string  `json:"company"`
	Location    *string  `json:"location"`
	URL         *string  `json:"url"`
	Description *string  `json:"description"`
	Skills      []string `json:"skills"`
	DatePosted  *string  `json:"date_posted"`
	IsRemote    *bool    `json:"is_remote"`
}

// JobSearchResponse is the public response for a transient JobTech search.
type JobSearchResponse struct {
	Count int               `json:"count"`
	Jobs  []JobSearchResult `json:"jobs"`
}

// ProfileJob records a job saved by a profile and its current application status.
type ProfileJob struct {
	ProfileID int       `json:"profile_id"`
	JobID     int       `json:"job_id"`
	Status    string    `json:"status"`
	SavedAt   time.Time `json:"saved_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SavedProfileJobs contains the jobs that have been saved for a specific profile.
type SavedProfileJob struct {
	JobID     int             `json:"job_id"`
	Job       JobSearchResult `json:"job"`
	Status    string          `json:"status"`
	SavedAt   time.Time       `json:"saved_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
