package ui

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/db"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/util"
	"github.com/gin-gonic/gin"
)

var _userStorer db.UserStorer
var _metaStorer db.MetadataStorer

var _cfg *config.Config

// InitUIRouteHandlers initializes
func InitUIRouteHandlers(cfg *config.Config, userStorer db.UserStorer, metaStorer db.MetadataStorer) {

	_cfg = cfg
	_userStorer = userStorer
	_metaStorer = metaStorer
}

// Register handles new registration requests
func Register(c *gin.Context) {
	var form loginForm
	if err := c.ShouldBindJSON(&form); err != nil {
		log.Error(err)
		util.BadReq(c, err.Error())
		return
	}

	// Check this user doesn't already exist
	users, err := _userStorer.GetUsers()
	if err != nil {
		log.Error(err)
		util.BadReq(c, err.Error())
		return
	}

	for _, u := range users {
		if u.Email == form.Email {
			util.BadReq(c, form.Email+" is already registered.")
			return
		}
	}

	var user messages.User
	user.Email = form.Email
	user.SetPassword(form.Password)
	user.GenId()

	err = _userStorer.RegisterUser(&user)
	if err != nil {
		log.Error(err)
		util.BadReq(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, user)
}

// Login handles login requests
func Login(c *gin.Context) {

	var form loginForm

	if err := c.ShouldBindJSON(&form); err != nil {
		log.Error(err)
		util.BadReq(c, err.Error())
		return
	}

	// Try to find the user
	users, err := _userStorer.GetUsers()
	if err != nil {
		log.Error(err)
		util.BadReq(c, err.Error())
		return
	}

	var user *messages.User
	for _, u := range users {
		if form.Email == u.Email {
			user = u
		}
	}

	if user == nil {
		log.Error(err)

		c.JSON(http.StatusUnauthorized, "Invalid email")
		return
	}

	if ok, err := user.CheckPassword(form.Password); err != nil || !ok {
		log.Error(err)
		c.JSON(http.StatusUnauthorized, "Invalid password")
		return
	}

	token := user.NewAuth0token("ui", "")
	tokenString, err := token.SignedString(_cfg.JWTSecretKey)
	if err != nil {
		util.BadReq(c, err.Error())
		return
	}

	c.String(http.StatusOK, tokenString)
}

// ListDocuments list all documents
func ListDocuments(c *gin.Context) {

	docs, err := _metaStorer.GetAllMetadata(false)
	if err != nil {
		log.Error(err)
		util.BadReq(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, docs)
}
