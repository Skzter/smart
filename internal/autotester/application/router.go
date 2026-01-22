package application

import (
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"

	ddgin "github.com/DataDog/dd-trace-go/contrib/gin-gonic/gin/v2"
	"github.com/gin-gonic/gin"

	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/internal/autotester/domain/handler"
	"gitlab.dit.htwk-leipzig.de/projekt2025-w-llm-unterstuetztes-autotesting-fuer-moderne-web-frontends/smart/web"
)

// NewRouter initializes the HTTP server, sets up API routes and serves static files.
// Registers endpoints and serves frontend assets from the embedded dist directory.
func NewRouter(logger *slog.Logger, controller *handler.AutotesterController, isHeadless bool) (*gin.Engine, error) {
	router := gin.Default()

	router.Use(gin.Recovery())
	router.Use(ddgin.Middleware(os.Getenv("DD_SERVICE")))

	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/chat", controller.HandleChatRequest)
		apiV1.POST("/validate", controller.HandleChatRequestValidity)
		apiV1.GET("/chats", controller.HandleGetChats)
		apiV1.GET("/chats/:chatId", controller.GetChatById)
		apiV1.GET("/template", controller.HandleGetTemplate)
		apiV1.POST("/saveLocal", controller.HandleSaveLocalRequest)
		apiV1.DELETE("/deleteLocal", controller.HandleDeleteLocalRequest)
		apiV1.POST("/run", controller.HandleRunContainer)
		apiV1.GET("/tests", controller.HandleGetRemoteTestcase)
		apiV1.GET("/test/:testId/stream", sseHeaderMiddleWare(), controller.HandleLogRequest)
		apiV1.POST("/auth/generate", internalOnlyMiddleware(logger), controller.HandleGenerateToken)

		apiV1.GET("/groups", controller.HandleGetGroups)
		apiV1.POST("/groups", controller.HandleCreateGroup)
		apiV1.POST("/chats/:chatId/groups", controller.HandleAssignChatToGroups)
		apiV1.DELETE("/chats/:chatId/groups/:groupId", controller.HandleRemoveChatFromGroup)
	}

	if !isHeadless {
		router.GET("/auth_config.json", func(c *gin.Context) {
			c.FileFromFS("/auth_config.json", http.FS(web.Auth0Config))
		})

		assetsFS, err := fs.Sub(web.DistFS, "dist/assets")
		if err != nil {
			return nil, err
		}
		router.StaticFS("/assets", http.FS(assetsFS))

		router.GET("/", func(c *gin.Context) {
			indexHTML, err := web.DistFS.ReadFile("dist/index.html")
			if err != nil {
				logger.Error("Failed to read embedded index.html", "error", err)
				c.String(http.StatusInternalServerError, "Internal Server Error")
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
		})
	}

	logger.Info("Router initialized", "isHeadless", isHeadless)
	return router, nil
}

func sseHeaderMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Next()
	}
}

// internalOnlyMiddleware restricts access to localhost and Docker internal networks.
// This prevents external access to sensitive endpoints like token generation.
func internalOnlyMiddleware(logger *slog.Logger) gin.HandlerFunc {
	// Define allowed CIDR ranges for Docker networks
	allowedCIDRs := []string{
		"127.0.0.0/8",    // Localhost IPv4
		"::1/128",        // Localhost IPv6
		"172.16.0.0/12",  // Docker default bridge
		"192.168.0.0/16", // Docker Compose networks
	}

	allowedNets := make([]*net.IPNet, 0, len(allowedCIDRs))
	for _, cidr := range allowedCIDRs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			logger.Error("Failed to parse CIDR", "cidr", cidr, "error", err)
			continue
		}
		allowedNets = append(allowedNets, ipNet)
	}

	return func(c *gin.Context) {
		clientIP := net.ParseIP(c.ClientIP())
		if clientIP == nil {
			logger.Warn("Invalid client IP", "ip", c.ClientIP())
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Forbidden: Invalid IP address",
			})
			return
		}

		// Check if client IP is in any allowed network
		for _, ipNet := range allowedNets {
			if ipNet.Contains(clientIP) {
				c.Next()
				return
			}
		}

		logger.Warn("Blocked external access attempt", "ip", clientIP.String(), "path", c.Request.URL.Path)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Forbidden: Endpoint only accessible from internal networks",
		})
	}
}
