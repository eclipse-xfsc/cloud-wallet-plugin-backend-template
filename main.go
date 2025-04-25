package main

import (
	"errors"
	"fmt"
	"net/http"
	"slices"

	pluginSdk "github.com/eclipse-xfsc/cloud-wallet-plugin-core"
	messaging "github.com/eclipse-xfsc/nats-message-library"
	"github.com/eclipse-xfsc/nats-message-library/common"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	cryptoTypes "gitlab.eclipse.org/eclipse/xfsc/libraries/crypto/engine/core/types"
)

const TenantId = "tenant_space"

func handlerApp(c *gin.Context) {
	c.JSON(http.StatusOK, `{"userId":1,"id":1,"title":"delectus aut autem","completed":false}`)
}

type eventPayload struct {
	Message string                    `json:"message"`
	Type    messaging.RecordEventType `json:"type"`
}

func handlerEvent(c *gin.Context) {
	topic := c.Param("topic")
	var data eventPayload
	err := c.BindJSON(&data)
	if !slices.Contains(messaging.RecordEventTypes(), data.Type) {
		c.JSON(http.StatusBadRequest, fmt.Sprintf("Invalid event type `%s`", data.Type))
		return
	}
	user := c.Request.Context().Value(pluginSdk.UserKey).(*pluginSdk.UserInfo)
	record := messaging.HistoryRecord{
		Reply: common.Reply{
			TenantId:  TenantId,
			RequestId: uuid.NewString(),
			Error:     nil,
		},
		UserId:  user.ID(),
		Message: data.Message,
	}
	err = pluginSdk.PublishHistoryEvent(topic, data.Type, record)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
	} else {
		c.JSON(http.StatusOK, record)
	}
}

func handlerCreateKey(c *gin.Context) {
	msg, err := pluginSdk.NewMessage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.Join(err, errors.New("failed to initialise keyMessenger")))
		return
	}
	user := c.Request.Context().Value(pluginSdk.UserKey).(*pluginSdk.UserInfo)
	keyId := uuid.NewString()
	err = msg.CreateKey(keyId, user.ID(), string(cryptoTypes.Ecdsap256))
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.Join(err, errors.New("failed to create key")))
		return
	}
	c.JSON(http.StatusOK, gin.H{"keyId": keyId})
}

func main() {
	r := gin.Default()
	r.GET("/app", pluginSdk.AuthMiddleware(), handlerApp)

	metadata := pluginSdk.Metadata{
		Name:        "plugin-template",
		ID:          "2392323923",
		Description: "Template of the plugin",
	}
	r.GET("/metadata", pluginSdk.MetadataHandler(metadata))
	r.POST("/event/:topic", pluginSdk.AuthMiddleware(), handlerEvent)
	r.POST("/key/create", pluginSdk.AuthMiddleware(), handlerCreateKey)
	r.Run()
}
