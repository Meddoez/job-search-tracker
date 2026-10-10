// Package routes registers the public HTTP API routes.
package routes

import (
	"net/http"

	handlers "github.com/Meddoez/job-search-tracker/backend/internal/handlers"
	"github.com/gin-gonic/gin"
)

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

// RegisterRoutes attaches the public API endpoints and their dependencies to router.
func RegisterRoutes(router *gin.Engine, client *http.Client, ingestionServiceURL string, profileRepo handlers.ProfileRepository, jobRepo handlers.ProfileJobRepository) {
	router.GET("/ping", ping)

	router.POST("/jobs/search", handlers.SearchJobsHandler(client, ingestionServiceURL))

	router.GET("/profiles", handlers.GetProfilesHandler(profileRepo))

	router.GET("/profiles/:id", handlers.GetProfileHandler(profileRepo))

	router.POST("/profiles", handlers.CreateProfileHandler(profileRepo))

	router.POST("/profiles/:id/jobs", handlers.SaveProfileJobHandler(jobRepo))

	router.PATCH("/profiles/:id/jobs/:jobId", handlers.UpdateProfileJobStatusHandler(jobRepo))

	router.DELETE("/profiles/:id", handlers.DeleteProfileHandler(profileRepo))

	router.GET("/profiles/:id/jobs", handlers.GetProfileJobsHandler(jobRepo))
}
