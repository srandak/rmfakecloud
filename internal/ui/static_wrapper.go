package ui

import (
	"net/http"
	"path"

	"github.com/ddvk/rmfakecloud/internal/webassets"
	"github.com/gin-gonic/gin"
)

const indexReplacement = "/default"

type ReactAppWrapper struct {
	fs     http.FileSystem
	prefix string
}

func NewReactAppWrapper(fs http.FileSystem, prefix string) *ReactAppWrapper {

	return &ReactAppWrapper{
		fs:     fs,
		prefix: "/static",
	}
}

func (w ReactAppWrapper) Open(filepath string) (http.File, error) {
	fullpath := filepath
	//index.html hack
	if filepath != indexReplacement {
		fullpath = path.Join(w.prefix, filepath)
	} else {
		fullpath = "/index.html"
	}
	f, err := w.fs.Open(fullpath)
	return f, err
}

func (w ReactAppWrapper) Register(router *gin.Engine) {

	router.StaticFS(w.prefix, w)

	//hack for index.html
	router.NoRoute(func(c *gin.Context) {
		c.FileFromFS(indexReplacement, w)
	})

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("/favicon.ico", webassets.Assets)
	})
}
