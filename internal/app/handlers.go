package app

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/email"
	"github.com/ddvk/rmfakecloud/internal/hwr"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/util"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func getHealth(c *gin.Context) {
	count := app.hub.ClientCount()
	c.String(http.StatusOK, "Working, %d clients", count)
}

func registerDevice(c *gin.Context) {

	var json messages.DeviceTokenRequest

	if err := c.ShouldBindJSON(&json); err != nil {
		util.BadReq(c, err.Error())
		return
	}

	log.Printf("Request: %s\n", json)

	// generate the JWT token
	expirationTime := time.Now().Add(356 * 24 * time.Hour)

	claims := &messages.Auth0token{
		DeviceDesc: json.DeviceDesc,
		DeviceId:   json.DeviceId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Subject:   "rM Device Token",
		},
	}

	deviceToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := deviceToken.SignedString(app.cfg.JWTSecretKey)
	if err != nil {
		util.BadReq(c, err.Error())
		return
	}

	c.String(200, tokenString)
}

func deleteDevice(c *gin.Context) {

	// TODO:
	c.String(204, "")
}

func requestDocument(c *gin.Context) {
	var req []messages.UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		util.BadReq(c, err.Error())
		return
	}

	response := []messages.UploadResponse{}

	for _, r := range req {
		id := r.Id
		if id == "" {
			util.BadReq(c, "no id")
		}
		url := app.docStorer.GetStorageURL(id)
		log.Debugln("StorageUrl: ", url)
		dr := messages.UploadResponse{BlobUrlPut: url, Id: id, Success: true, Version: r.Version}
		response = append(response, dr)
	}

	c.JSON(http.StatusOK, response)
}

func updateDocumentStatus(c *gin.Context) {
	var req []messages.RawDocument
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error(err)
		util.BadReq(c, err.Error())
		return
	}
	result := []messages.StatusResponse{}
	for _, r := range req {
		log.Println("For Id: ", r.Id)
		log.Println(" Name: ", r.VissibleName)

		ok := false
		event := "DocAdded"
		message := ""

		err := app.metaStorer.UpdateMetadata(&r)
		if err == nil {
			ok = true
			//fix it: id of subscriber
			msg := newWs(&r, event)
			app.hub.Send(msg)
		} else {
			message = err.Error()
			log.Error(err)
		}
		result = append(result, messages.StatusResponse{Id: r.Id, Success: ok, Message: message})
	}

	c.JSON(http.StatusOK, result)
}

func createAuthenticationToken(c *gin.Context) {

	deviceToken, err := getToken(c, app.cfg.JWTSecretKey)

	if err != nil {
		log.Warnln(err)
	}

	log.Debug(deviceToken)

	expirationTime := time.Now().Add(30 * 24 * time.Hour)

	claims := &messages.Auth0token{
		Profile: &messages.Auth0profile{
			UserId:        "auth0|1234",
			IsSocial:      false,
			Name:          "rmFake",
			Nickname:      "rmFake",
			Email:         "fake@rmfake",
			EmailVerified: true,
			Picture:       "image.png",
			CreatedAt:     time.Date(2020, 04, 29, 10, 48, 25, 936, time.UTC),
			UpdatedAt:     time.Now(),
		},
		DeviceDesc: deviceToken.DeviceDesc,
		DeviceId:   deviceToken.DeviceId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			Subject:   "rM User Token",
		},
	}

	userToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := userToken.SignedString(app.cfg.JWTSecretKey)

	if err != nil {
		util.BadReq(c, err.Error())
		return
	}
	//TODO: do something with the token

	c.String(200, tokenString)
}

func getServiceLocator(c *gin.Context) {
	svc := c.Param("service")
	log.Printf("Requested: %s\n", svc)
	response := messages.HostResponse{Host: config.DefaultHost, Status: "OK"}
	c.JSON(http.StatusOK, response)
}

func getDocuments(c *gin.Context) {

	withBlob, _ := strconv.ParseBool(c.Query("withBlob"))
	docID := c.Query("doc")
	log.Println("params: withBlob, docId", withBlob, docID)
	result := []*messages.RawDocument{}

	var err error
	if docID != "" {
		//load single document
		var doc *messages.RawDocument
		doc, err = app.metaStorer.GetMetadata(docID, withBlob)
		if err == nil {
			result = append(result, doc)
		}
	} else {
		//load all
		result, err = app.metaStorer.GetAllMetadata(withBlob)
	}

	if err != nil {
		log.Error(err)
		util.InternalError(c, "blah")
		return
	}

	c.JSON(http.StatusOK, result)
}

func deleteDocument(c *gin.Context) {
	var req []messages.IdRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("bad request")
		util.BadReq(c, err.Error())
		return
	}

	result := []messages.StatusResponse{}
	for _, r := range req {
		metadata, err := app.metaStorer.GetMetadata(r.Id, false)
		ok := true
		if err == nil {
			err := app.docStorer.RemoveDocument(r.Id)
			if err != nil {
				log.Error(err)
				ok = false
			}
			msg := newWs(metadata, "DocDeleted")
			app.hub.Send(msg)
		}
		result = append(result, messages.StatusResponse{Id: r.Id, Success: ok})
	}

	c.JSON(http.StatusOK, result)
}

func sendEmail(c *gin.Context) {
	log.Println("Sending email")

	form, err := c.MultipartForm()
	if err != nil {
		log.Error(err)
		util.InternalError(c, "not multipart form")
		return
	}
	for k := range form.File {
		log.Debugln("form field", k)
	}
	for k := range form.Value {
		log.Debugln("form value", k)
	}

	emailClient := email.EmailBuilder{
		Subject: form.Value["subject"][0],
		ReplyTo: form.Value["reply-to"][0],
		From:    form.Value["from"][0],
		To:      form.Value["to"][0],
		Body:    util.StripAds(form.Value["html"][0]),
	}

	for _, file := range form.File["attachment"] {
		f, err := file.Open()
		defer f.Close()
		if err != nil {
			log.Error(err)
			util.InternalError(c, "cant open attachment")
			return
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			log.Error(err)
			util.InternalError(c, "cant read attachment")
			return
		}
		emailClient.AddFile(file.Filename, data, file.Header.Get("Content-Type"))
	}
	err = emailClient.Send()
	if err != nil {
		log.Error(err)
		util.InternalError(c, "cant send email")
		return
	}
	c.String(http.StatusOK, "")
}

func getHwr(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil || len(body) < 1 {
		log.Warn("no body")
		util.BadReq(c, "missing bbody")
		return
	}
	response, err := hwr.SendRequest(body)
	if err != nil {
		log.Error(err)
		util.InternalError(c, "cannot send")
		return
	}
	c.Data(http.StatusOK, hwr.JIIX, response)

}

func getNotifications(c *gin.Context) {
	userID := c.GetString("userId")
	log.Println("accepting websocket", userID)
	app.hub.ConnectWs(c.Writer, c.Request)
	log.Println("closing the ws")
}

// TODO: (fs *Storage)
func getStorage(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.String(400, "set up us the bomb")
		return
	}

	//todo: storage provider
	log.Printf("Requestng Id: %s\n", id)

	reader, err := app.docStorer.GetDocument(id)
	defer reader.Close()

	if err != nil {
		log.Error(err)
		c.String(500, "internal error")
		c.Abort()
		return
	}

	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

func putStorage(c *gin.Context) {
	id := c.Query("id")
	log.Printf("Uploading id %s\n", id)
	body := c.Request.Body
	defer body.Close()

	err := app.docStorer.StoreDocument(body, id)
	if err != nil {
		log.Error(err)
		c.String(500, "set up us the bomb")
		c.Abort()
		return
	}

	c.JSON(200, gin.H{})
}
