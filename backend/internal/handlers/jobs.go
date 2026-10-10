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
	GetProfileJobs(
		ctx context.Context,
		porfileID int) ([]models.SavedProfileJob, error)
}

type StatusRequest struct {
	Status string `json:"status" binding:"required"`
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

func UpdateProfileJobStatusHandler(repo ProfileJobRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// should we dedup commands like this? they are used a lot of the time? or does it make it more unclear?
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

		var request StatusRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "A valid status is required",
			})
			return
		}

		switch request.Status {
		case "saved", "applied", "interviewing", "offer", "rejected", "withdrawn":
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "A valid job status is required",
			})
			return
		}

		profileJob, err := repo.UpdateProfileJobStatus(c.Request.Context(), profileID, jobID, request.Status)
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Saved job not found",
			})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to update profile job status",
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
