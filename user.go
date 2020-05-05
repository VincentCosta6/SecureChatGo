package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	options2 "go.mongodb.org/mongo-driver/mongo/options"
)

type FindLikeUsersStruct struct {
	LikeName string
}

func FindLikeUsers(c *gin.Context) {
	LikeName := c.Param("LikeName")

	if LikeName == "" {
		c.JSON(400, gin.H{"message": "You must send a username"})
		return
	}

	cursor, err := Users.Find(context.TODO(), bson.D{{"username", primitive.Regex{Pattern: LikeName, Options: "i"}}})

	users := make([]UserSchema, 0)

	if err != nil {
		c.JSON(500, gin.H{"message": "Could not find like users", "err": err})
		return
	}

	for cursor.Next(context.Background()) {

		user := UserSchema{}
		err := cursor.Decode(&user)
		if err != nil {
			//handle err
		}

		user.Password = ""

		users = append(users, user)
	}

	c.JSON(200, gin.H{"message": "Successful", "results": users})
}

func FindChannels(c *gin.Context) {
	userContext, exists := c.Get("user")

	if !exists {
		fmt.Println(exists)
	}

	user := userContext.(UserSchema)

	key := "privatekeys." + user.ID.Hex()

	cursor, err := Channels.Find(context.TODO(), bson.M{key: bson.M{"$exists": true}})

	channels := make([]ChannelSchema, 0)

	if err != nil {
		c.JSON(500, gin.H{"message": "Could not find channels", "err": err})
		return
	}

	for cursor.Next(context.Background()) {

		channel := ChannelSchema{}
		err := cursor.Decode(&channel)
		if err != nil {
			//handle err
			fmt.Println("err", err)
		}

		channels = append(channels, channel)
	}

	completeChannels := make([]FindChannelsType, 0)

	for _, channel := range channels {
		options := options2.Find()

		options.SetSort(bson.M{"_id": 1})

		options.SetLimit(50)

		messagesCursor, err := Messages.Find(context.TODO(), bson.M{"channelid": channel.ID}, options)

		if err != nil {
			fmt.Println("No messages in channel", channel.ID)
		}

		messages := make([]MessageSchema, 0)

		for messagesCursor.Next(context.Background()) {
			message := MessageSchema{}
			err := messagesCursor.Decode(&message)

			if err != nil {
				fmt.Println("err", err)
			}

			messages = append(messages, message)
		}

		completeChannels = append(completeChannels, FindChannelsType{channel.ID, channel.Name, channel.PrivateKeys, channel.UserMap, messages})
	}

	c.JSON(200, gin.H{"results": completeChannels})
}

type FindChannelsType struct {
	ID primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	Name string
	PrivateKeys map[string]string // [userID]: Channels symmetric AES key is encrypted with the select users public key
	UserMap map[string]string
	Messages []MessageSchema
}