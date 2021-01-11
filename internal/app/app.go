package app

import (
	"context"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/db"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/ddvk/rmfakecloud/internal/ui"
	"github.com/gin-gonic/gin"
)

// App web app
type App struct {
	router     *gin.Engine
	cfg        *config.Config
	srv        *http.Server
	hub        *Hub
	metaStorer db.MetadataStorer
	docStorer  storage.DocumentStorer
	userStorer db.UserStorer
}

var app App

// NewApp constructs an app
func NewApp(cfg *config.Config, metaStorer db.MetadataStorer, docStorer storage.DocumentStorer, userStorer db.UserStorer) App {

	hub := NewHub()

	gin.ForceConsoleColor()
	router := gin.Default()

	ui.InitUIRouteHandlers(cfg, userStorer, metaStorer)

	app = App{
		router:     router,
		cfg:        cfg,
		hub:        hub,
		docStorer:  docStorer,
		userStorer: userStorer,
		metaStorer: metaStorer,
	}

	//TODO: refactor
	initializeRoutes(router, cfg, hub)

	return app
}

// Start starts the app
func (app *App) Start() {
	app.srv = &http.Server{
		Addr:    ":" + app.cfg.Port,
		Handler: app.router,
	}

	if err := app.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}

// Stop shuts down the app serv
func (app *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
}
