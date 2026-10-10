package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Meddoez/job-search-tracker/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// ProfileRequest describes the JSON payload accepted when creating a profile.
type ProfileRequest struct {
	Name   string   `json:"name" binding:"required"`
	Skills []string `json:"skills"`
}

// ProfileRepository defines the profile operations required by the HTTP handlers.
type ProfileRepository interface {
	// CreateProfile stores a profile and its skills.
	CreateProfile(ctx context.Context, name string, skills []string) (models.Profile, error)
	// GetProfile returns the profile with the given ID.
	GetProfile(ctx context.Context, id int) (models.Profile, error)
	// GetProfiles returns all stored profiles.
	GetProfiles(ctx context.Context) ([]models.Profile, error)
	// DeleteProfile deletes the profile with the specific iD.
	DeleteProfile(ctx context.Context, id int) error
}

// CreateProfileHandler returns an HTTP handler that validates and creates profiles.
// Duplicate normalized names produce a 409 Conflict response.
func CreateProfileHandler(repo ProfileRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request ProfileRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "A profile name is required",
			})
			return
		}
		request.Name = strings.TrimSpace(request.Name)
		if request.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "A profile name is required",
			})
			return
		}

		profile, err := repo.CreateProfile(c.Request.Context(), request.Name, request.Skills)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "uq_profiles_name_normalized" {
				c.JSON(http.StatusConflict, gin.H{
					"message": "A profile with this name already exists",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to create profile",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"profile": profile,
		})
	}
}

// GetProfilesHandler returns an HTTP handler that lists all stored profiles.
func GetProfilesHandler(repo ProfileRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		profiles, err := repo.GetProfiles(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to get profiles",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"profiles": profiles,
		})
	}
}

// GetProfileHandler returns an HTTP handler that retrieves a profile by ID.
func GetProfileHandler(repo ProfileRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil || id <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "A valid profile ID is required",
			})
			return
		}

		profile, err := repo.GetProfile(c.Request.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Profile not found",
			})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to get profile",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"profile": profile,
		})
	}
}

// DeleteProfileHandler returns an HTTP handler that deletes a profile by ID.
func DeleteProfileHandler(repo ProfileRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		profileID, err := strconv.Atoi(c.Param("id"))
		if err != nil || profileID <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "A valid profile ID is required",
			})
			return
		}

		err = repo.DeleteProfile(c.Request.Context(), profileID)
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Profile not found",
			})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to delete profile",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Profile successfully deleted",
		})
	}
}
