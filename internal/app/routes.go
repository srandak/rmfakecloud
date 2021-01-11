package app

import (
	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/gin-gonic/gin"

	"github.com/ddvk/rmfakecloud/internal/ui"
	"github.com/ddvk/rmfakecloud/internal/webassets"

	log "github.com/sirupsen/logrus"
)

func initializeRoutes(router *gin.Engine, cfg *config.Config, hub *Hub) {

	// router.Use(ginlogrus.Logger(std.Out), gin.Recovery())

	if log.GetLevel() == log.DebugLevel {
		router.Use(requestLoggerMiddleware())
	}

	router.Use(requestLoggerMiddleware())

	ui.NewReactAppWrapper(webassets.Assets, "/static").Register(router)

	r := router.Group("/ui/api")
	{
		r.POST("register", ui.Register)
		r.POST("login", ui.Login)

		// UI Authenticated Only TODO: ui/api
		r.GET("list", authMiddleware(cfg.JWTSecretKey), ui.ListDocuments)
	}

	router.GET("/health", getHealth)

	// register device
	router.POST("/token/json/2/device/new", registerDevice)

	// create new access token
	router.POST("/token/json/2/user/new", createAuthenticationToken)

	//service locator
	router.GET("/service/json/1/:service", getServiceLocator)

	auth := router.Group("/")

	auth.Use(authMiddleware(cfg.JWTSecretKey))
	{
		auth.GET("storage", getStorage)

		//todo: pass the token in the url
		auth.PUT("storage", putStorage)

		//unregister device
		auth.POST("token/json/3/device/delete", deleteDevice)

		// websocket notifications
		auth.GET("notifications/ws/json/1", getNotifications)
		// live sync
		auth.GET("livesync/ws/json/2/:authid/sub", func(c *gin.Context) {
			hub.ConnectWs(c.Writer, c.Request)
		})

		auth.PUT("document-storage/json/2/upload/request", requestDocument)
		auth.PUT("document-storage/json/2/upload/update-status", updateDocumentStatus)
		auth.PUT("document-storage/json/2/delete", deleteDocument)
		auth.GET("document-storage/json/2/docs", getDocuments)

		// send email
		auth.POST("api/v2/document", sendEmail)

		// hwr
		auth.POST("api/v1/page", getHwr)
	}
}
