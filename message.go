package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type CreateMessageType struct {
	ChannelID string
	Message string
}

func CreateMessageRoute(c *gin.Context) {
	var form CreateMessageType

	c.BindJSON(&form)

	if form.ChannelID == "" {
		c.JSON(400, gin.H{"message": "You must send a channel id"})
		return
	}

	if form.Message == "" {
		c.JSON(400, gin.H{"message": "You must send a message"})
		return
	}

	convertID, err := primitive.ObjectIDFromHex(form.ChannelID)

	if err != nil {
		c.JSON(500, gin.H{"message": "Error converting your id to a mongo id", "err": err})
		return
	}

	var foundChannel ChannelSchema

	fmt.Println(convertID)

	err = Channels.FindOne(context.TODO(), bson.D{{"_id", convertID}}).Decode(&foundChannel)

	if err != nil {
		c.JSON(400, gin.H{ "message": "Channel does not exist" })
		return
	}

	newMessage := MessageSchema{primitive.NewObjectID(), convertID, time.Now(), form.Message}

	_, err = Messages.InsertOne(context.TODO(), newMessage)

	if err != nil {
		c.JSON(400, gin.H{"message": "Could not insert message", "err": err})
		return
	}

	c.JSON(200, gin.H{"message": "Message sent successfully", "inserted": newMessage})
}