package app

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/util"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func authMiddleware(jwtSecretKey []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := getToken(c, jwtSecretKey)
		if err == nil {
			c.Set("userId", strings.TrimPrefix(claims.Profile.UserId, "auth0|"))
			log.Info("got a user token", claims.Profile.UserId)
		} else {
			log.Warn(err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
		}
		c.Next()
	}
}

func getToken(c *gin.Context, jwtSecretKey []byte) (claims *messages.Auth0Token, err error) {
	auth := c.Request.Header["Authorization"]

	if len(auth) < 1 {
		util.AccessDenied(c, "missing token")
		return nil, errors.New("missing token")
	}

	token := strings.Split(auth[0], " ")

	if len(token) < 2 {
		return nil, errors.New("missing token")
	}

	claims = &messages.Auth0Token{}
	_, err = jwt.ParseWithClaims(token[1], claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})
	return
}

var ignored = []string{"/storage", "/api/v2/document"}

func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		log.Debugln("header ", c.Request.Header)
		for _, skip := range ignored {
			if skip == c.Request.URL.Path {
				log.Debugln("body logging ignored")
				c.Next()
				return
			}
		}

		var buf bytes.Buffer
		tee := io.TeeReader(c.Request.Body, &buf)
		body, _ := ioutil.ReadAll(tee)
		c.Request.Body = ioutil.NopCloser(&buf)
		log.Debugln("body: ", string(body))
		c.Next()
	}
}
