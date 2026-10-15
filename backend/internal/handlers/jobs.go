// Package handlers implements HTTP handlers for the public API.
package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Meddoez/job-search-tracker/backend/internal/models"
	"github.com/gin-gonic/gin"
)

type ProfileJobRepository interface {
	SaveProfileJob(
		ctx context.Context,
		profileID int,
		jobs models.JobSearchResult,
	) (models.ProfileJob, error)
	UpdateProfileJobStatus(
		ctx context.Context,
		profileID int,
		jobID int,
		status string,
	) (models.ProfileJob, error)
	UpdateProfileJobNotes(
		ctx context.Context,
		profileID int,
		jobID int,
		notes string,
	) (models.ProfileJob, error)
	GetProfileJobs(
		ctx context.Context,
		porfileID int) ([]models.SavedProfileJob, error)
}

type UpdateProfileJobRequest struct {
	Status *string `json:"status"`
	Notes  *string `json:"notes"`
}

// SearchJobsHandler proxies transient searches to ingestion and validates its response.
func SearchJobsHandler(client *http.Client, ingestionServiceURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		request, err := http.NewRequestWithContext(
			c.Request.Context(),
			http.MethodPost,
			ingestionServiceURL+"/search",
			c.Request.Body,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to create search request",
			})
			return
		}
		request.Header.Set("Content-Type", "application/json")

		response, err := client.Do(request)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"message": "Job search service unavailable",
			})
			return
		}
		defer response.Body.Close()

		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			c.DataFromReader(
				response.StatusCode,
				response.ContentLength,
				response.Header.Get("Content-Type"),
				response.Body,
				nil,
			)
			return
		}

		var result models.JobSearchResponse
		if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"message": "Job search service returned an invalid response",
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func SaveProfileJobHandler(repo ProfileJobRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		profileID, err := strconv.Atoi(c.Param("id"))
		if err != nil || profileID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid profile ID",
			})
			return
		}

		var job models.JobSearchResult
		if err := c.ShouldBindBodyWithJSON(&job); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "A valid Job ID is required",
			})
			return
		}

		if job.Source != "jobtech" ||
			job.ExternalID == nil ||
			strings.TrimSpace(*job.ExternalID) == "" ||
			strings.TrimSpace(job.Title) == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Job source, external ID, and title are required",
			})
			return
		}

		profileJob, err := repo.SaveProfileJob(c.Request.Context(), profileID, job)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to save profile job",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"profile_job": profileJob,
		})
	}
}

// TODO: Check if its possible to combine UpdateProfileJobsNotes and Status into a single function.
func UpdateProfileJobHandler(repo ProfileJobRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		profileID, err := strconv.Atoi(c.Param("id"))
		if err != nil || profileID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "A valid Profile ID is required",
			})
			return
		}

		jobID, err := strconv.Atoi(c.Param("jobId"))
		if err != nil || jobID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "A valid job ID is required",
			})
			return
		}

		var request UpdateProfileJobRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "A valid update request is required",
			})
			return
		}

		if request.Status == nil && request.Notes == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "At least one of status or notes is required",
			})
			return
		}

		// If the Status request is invalid, return and dont update notes either.
		var profileJob models.ProfileJob
		if request.Status != nil {
			switch *request.Status {
			case "saved", "applied", "interviewing", "offer", "rejected", "withdrawn":
				profileJob, err = repo.UpdateProfileJobStatus(c.Request.Context(), profileID, jobID, *request.Status)
				if errors.Is(err, sql.ErrNoRows) {
					c.JSON(http.StatusNotFound, gin.H{
						"message": "Saved job not found",
					})
					return
				}
			default:
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "A valid job status is required",
				})
				return
			}
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to update profile job status",
			})
			return
		}

		if request.Notes != nil {
			profileJob, err = repo.UpdateProfileJobNotes(c.Request.Context(), profileID, jobID, *request.Notes)
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{
					"message": "Saved jobs not found",
				})
				return
			}
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to update profile job notes",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"profile_job": profileJob,
		})

	}
}

func GetProfileJobsHandler(repo ProfileJobRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		profileID, err := strconv.Atoi(c.Param("id"))
		if err != nil || profileID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "A valid profileID is required",
			})
			return
		}

		savedProfileJobs, err := repo.GetProfileJobs(c.Request.Context(), profileID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to get profile jobs",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"profile_job": savedProfileJobs,
		})
	}
}

func DeleteProfileJobHandler(repo ProfileJobRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		profileID, err := strconv.Atoi(c.Param("id"))
		if err != nil  || profileID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "A valid profileID is required"
			})
		}

		jobID, err := strconv.Atoi(c.Param("jobId"))		
		if err != nil || jobID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "A valid jobID is required"
			})
		}

		
	}
}