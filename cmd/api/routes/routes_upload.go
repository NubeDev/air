package routes

import (
	"github.com/NubeDev/air/cmd/api/handlers/upload"
	"github.com/gin-gonic/gin"
)

// SetupUploadRoutes configures file upload routes
func SetupUploadRoutes(rg *gin.RouterGroup) {
	uploadGroup := rg.Group("/upload")
	{
		uploadGroup.POST("/file", upload.UploadFile())
		uploadGroup.GET("/files", upload.ListUploadedFiles())
		uploadGroup.GET("/file/:id", upload.GetUploadedFile())
		uploadGroup.DELETE("/file/:id", upload.DeleteUploadedFile())
		uploadGroup.POST("/file/:id/learn", upload.LearnFileSchema())
	}
}
