package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleGetScreenshot handles requests for test screenshots.
// GET /api/v1/test/:testId/screenshot
// Returns a JSON response with the S3 URL to the screenshot.
func (a *AutotesterController) HandleGetScreenshot(c *gin.Context) {
	ctx, span := a.tracer.Start(c, "autotesterController.HandleGetScreenshot")
	defer span.End()

	testID := c.Param("testId")
	if testID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "testId is required"})
		return
	}

	// Check if screenshot exists
	hasScreenshot, _ := a.mediaStorageService.HasMedia(ctx, testID)
	if !hasScreenshot {
		c.JSON(http.StatusNotFound, gin.H{"error": "no screenshot found for this test"})
		return
	}

	// Get the screenshot URL from S3
	screenshotUrl, err := a.mediaStorageService.GetScreenshotUrl(ctx, testID)
	if err != nil {
		a.logger.Error("failed to get screenshot URL",
			"testId", testID,
			"error", err.Error(),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve screenshot"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"testId": testID,
		"url":    screenshotUrl,
	})
}

// HandleGetVideo handles requests for test videos.
// GET /api/v1/test/:testId/video
// Returns a JSON response with the S3 URL to the video.
func (a *AutotesterController) HandleGetVideo(c *gin.Context) {
	ctx, span := a.tracer.Start(c, "autotesterController.HandleGetVideo")
	defer span.End()

	testID := c.Param("testId")
	if testID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "testId is required"})
		return
	}

	// Check if video exists
	_, hasVideo := a.mediaStorageService.HasMedia(ctx, testID)
	if !hasVideo {
		c.JSON(http.StatusNotFound, gin.H{"error": "no video found for this test"})
		return
	}

	// Get the video URL from S3
	videoUrl, err := a.mediaStorageService.GetVideoUrl(ctx, testID)
	if err != nil {
		a.logger.Error("failed to get video URL",
			"testId", testID,
			"error", err.Error(),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve video"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"testId": testID,
		"url":    videoUrl,
	})
}

// HandleGetMediaInfo handles requests to check if media exists for a test.
// GET /api/v1/test/:testId/media
func (a *AutotesterController) HandleGetMediaInfo(c *gin.Context) {
	ctx, span := a.tracer.Start(c, "autotesterController.HandleGetMediaInfo")
	defer span.End()

	testID := c.Param("testId")
	if testID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "testId is required"})
		return
	}

	hasScreenshot, hasVideo := a.mediaStorageService.HasMedia(ctx, testID)

	c.JSON(http.StatusOK, gin.H{
		"testId":        testID,
		"hasScreenshot": hasScreenshot,
		"hasVideo":      hasVideo,
	})
}
