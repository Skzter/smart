package handler

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

// HandleGetScreenshot handles requests for test screenshots.
// GET /api/v1/test/:testId/screenshot
func (a *AutotesterController) HandleGetScreenshot(c *gin.Context) {
	_, span := a.tracer.Start(c, "autotesterController.HandleGetScreenshot")
	defer span.End()

	testID := c.Param("testId")
	if testID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "testId is required"})
		return
	}

	// Check if screenshot exists
	hasScreenshot, _ := a.mediaStorageService.HasMedia(testID)
	if !hasScreenshot {
		c.JSON(http.StatusNotFound, gin.H{"error": "no screenshot found for this test"})
		return
	}

	// Get the screenshot path for serving via file
	screenshotPath, err := a.mediaStorageService.GetScreenshotPath(testID)
	if err != nil {
		a.logger.Error("failed to get screenshot path",
			"testId", testID,
			"error", err.Error(),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve screenshot"})
		return
	}

	// Serve the file with proper content type
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "public, max-age=3600")
	c.File(screenshotPath)
}

// HandleGetVideo handles requests for test videos.
// GET /api/v1/test/:testId/video
func (a *AutotesterController) HandleGetVideo(c *gin.Context) {
	_, span := a.tracer.Start(c, "autotesterController.HandleGetVideo")
	defer span.End()

	testID := c.Param("testId")
	if testID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "testId is required"})
		return
	}

	// Check if video exists
	_, hasVideo := a.mediaStorageService.HasMedia(testID)
	if !hasVideo {
		c.JSON(http.StatusNotFound, gin.H{"error": "no video found for this test"})
		return
	}

	// Get the video path
	videoPath, err := a.mediaStorageService.GetVideoPath(testID)
	if err != nil {
		a.logger.Error("failed to get video path",
			"testId", testID,
			"error", err.Error(),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve video"})
		return
	}

	// Get file info for content length
	fileInfo, err := os.Stat(videoPath)
	if err != nil {
		a.logger.Error("failed to stat video file",
			"testId", testID,
			"error", err.Error(),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve video info"})
		return
	}

	// Set headers for video streaming
	c.Header("Content-Type", "video/webm")
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=3600")

	// Serve the file (Gin handles range requests automatically)
	c.File(videoPath)
}

// HandleGetMediaInfo handles requests to check if media exists for a test.
// GET /api/v1/test/:testId/media
func (a *AutotesterController) HandleGetMediaInfo(c *gin.Context) {
	_, span := a.tracer.Start(c, "autotesterController.HandleGetMediaInfo")
	defer span.End()

	testID := c.Param("testId")
	if testID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "testId is required"})
		return
	}

	hasScreenshot, hasVideo := a.mediaStorageService.HasMedia(testID)

	c.JSON(http.StatusOK, gin.H{
		"testId":        testID,
		"hasScreenshot": hasScreenshot,
		"hasVideo":      hasVideo,
	})
}
