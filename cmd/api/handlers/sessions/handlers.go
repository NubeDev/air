package sessions

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/store"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// StartSession starts a new learning session
func StartSession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req store.StartSessionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		// Convert options to JSON string
		optionsJSON, err := json.Marshal(req.Options)
		if err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid options",
				Details: err.Error(),
			})
			return
		}

		// Create session
		session := store.Session{
			Name:           req.SessionName,
			FilePath:       req.FilePath,
			Status:         "active",
			DatasourceType: req.DatasourceType,
			Options:        string(optionsJSON),
		}

		if err := db.Create(&session).Error; err != nil {
			logger.LogError(logger.ServiceREST, "Failed to create session", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to create session",
				Details: err.Error(),
			})
			return
		}

		logger.LogInfo(logger.ServiceREST, "Session created", map[string]interface{}{
			"session_id": session.ID,
			"name":       session.Name,
			"file_path":  session.FilePath,
		})

		c.JSON(http.StatusCreated, session)
	}
}

// GetSession retrieves a session by ID
func GetSession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid session ID",
				Details: err.Error(),
			})
			return
		}

		var session store.Session
		if err := db.First(&session, uint(id)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, store.ErrorResponse{
					Error: "Session not found",
				})
				return
			}
			logger.LogError(logger.ServiceREST, "Failed to get session", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to get session",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, session)
	}
}

// ListSessions lists all sessions
func ListSessions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sessions []store.Session
		if err := db.Order("created_at DESC").Find(&sessions).Error; err != nil {
			logger.LogError(logger.ServiceREST, "Failed to list sessions", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to list sessions",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"sessions": sessions,
		})
	}
}

// GetSessionStatus gets the status of a session
func GetSessionStatus(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid session ID",
				Details: err.Error(),
			})
			return
		}

		var session store.Session
		if err := db.Select("id, status, created_at, updated_at").First(&session, uint(id)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, store.ErrorResponse{
					Error: "Session not found",
				})
				return
			}
			logger.LogError(logger.ServiceREST, "Failed to get session status", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to get session status",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"session_id": session.ID,
			"status":     session.Status,
			"created_at": session.CreatedAt,
			"updated_at": session.UpdatedAt,
		})
	}
}

// EndSession ends a learning session
func EndSession(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid session ID",
				Details: err.Error(),
			})
			return
		}

		// Update session status to completed
		result := db.Model(&store.Session{}).Where("id = ?", uint(id)).Update("status", "completed")
		if result.Error != nil {
			logger.LogError(logger.ServiceREST, "Failed to end session", result.Error)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to end session",
				Details: result.Error.Error(),
			})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, store.ErrorResponse{
				Error: "Session not found",
			})
			return
		}

		logger.LogInfo(logger.ServiceREST, "Session ended", map[string]interface{}{
			"session_id": id,
		})

		c.JSON(http.StatusOK, store.SuccessResponse{
			Message: "Session ended successfully",
		})
	}
}
