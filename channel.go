package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateType struct {
	Name string
	PrivateKeys map[string]string
}

func CreateChannelRoute(c *gin.Context) {
	var form CreateType

	c.BindJSON(&form)

	if form.Name == "" {
		c.JSON(400, gin.H{"message": "You must send a name of the channel"})
		return
	}

	if len(form.PrivateKeys) == 0 {
		c.JSON(400, gin.H{"message": "You must send a map of the privateKeys and who they belong to"})
		return
	}

	newChannel := ChannelSchema{primitive.NewObjectID(), form.Name, form.PrivateKeys}

	go(func() {
		message := WebsocketMessageType{"NEW_CHANNEL", newChannel}

		clients := make([]string, 0)

		for key, _ := range form.PrivateKeys {
			clients = append(clients, key)
		}

		HubGlob.createMessage <- CreatedMessageStruct{message: &message, clients: &clients}
	})()

	_, err := Channels.InsertOne(context.TODO(), newChannel)

	if err != nil {
		fmt.Println("Fatal error while creating channel:", err)
		c.JSON(500, gin.H{"message": "Failed to create channel", "err": err})
		return
	}

	c.JSON(200, gin.H{"message": "Channel has been created", "channel": newChannel})
}