package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CreateType struct {
	Name string
	PrivateKeys map[string]string
	UserMap map[string]string
}

type AddType struct {
	ChannelID string
	PrivateKeys map[string]string
	UserMap map[string]string
}

type LeaveChannel struct {
	ChannelID string
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

	if len(form.UserMap) == 0 {
		c.JSON(400, gin.H{"message": "You must send a map of the user_ids and usernames"})
		return
	}

	newChannel := ChannelSchema{primitive.NewObjectID(), form.Name, form.PrivateKeys, form.UserMap}

	go(func() {
		message := WebsocketMessageType{"NEW_CHANNEL", newChannel}

		clients := make([]string, 0)

		for key := range form.PrivateKeys {
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

func AddUserRoute(c *gin.Context) {
	var form AddType

	c.BindJSON(&form)

	if form.ChannelID == "" {
		c.JSON(400, gin.H{"message": "You must send the id of the channel"})
		fmt.Println("No id")
		return
	}

	if len(form.PrivateKeys) == 0 {
		c.JSON(400, gin.H{"message": "You must send a map of the privateKeys and who they belong to"})
		fmt.Println("No keys")
		return
	}

	if len(form.UserMap) == 0 {
		c.JSON(400, gin.H{"message": "You must send a map of the user_ids and usernames"})
		return
	}

	id, err := primitive.ObjectIDFromHex(form.ChannelID)

	if err != nil {
		c.JSON(400, gin.H{"message": "Id is invalid and not a Mongo ID"})
		fmt.Println("No id")
		return
	}

	var channel ChannelSchema

	err = Channels.FindOne(context.TODO(), bson.D{{ "_id", id }}).Decode(&channel)

	if err != nil {
		c.JSON(400, gin.H{"message": "Couldnt find channel"})
		fmt.Println("No id")
		return
	}

	addUser := AddUser{ChannelID:id, NewUsers:form.PrivateKeys, UserMap:form.UserMap}

	go(func() {
		message := WebsocketMessageType{"ADD_USER", addUser}

		clients := make([]string, 0)

		for key := range form.PrivateKeys {
			clients = append(clients, key)
		}

		HubGlob.createMessage <- CreatedMessageStruct{message: &message, clients: &clients}
	})()

	converted := make(map[string]string)

	for k, v := range form.PrivateKeys {
		converted["privatekeys." + k] = v
	}

	for k, v := range form.UserMap {
		converted["usermap." + k] = v
	}

	_, err = Channels.UpdateOne(context.TODO(), bson.D{{ "_id", id }}, bson.D{{ "$set", converted }})

	if err != nil {
		fmt.Println(err)
		c.JSON(500, gin.H{"message": "Error", "err": err})
		return
	}

	c.JSON(200, gin.H{"message": "Users have been added"})
}

func LeaveChannelRoute(c *gin.Context) {
	var form LeaveChannel

	c.BindJSON(&form)

	if form.ChannelID == "" {
		c.JSON(400, gin.H{"message": "You must send a channel id"})
		return
	}

	convertID, err := primitive.ObjectIDFromHex(form.ChannelID)

	if err != nil {
		c.JSON(500, gin.H{"message": "Error converting your channel id to a mongo id", "err": err})
		return
	}

	var foundChannel ChannelSchema

	err = Channels.FindOne(context.TODO(), bson.D{{"_id", convertID}}).Decode(&foundChannel)

	if err != nil {
		c.JSON(400, gin.H{ "message": "Channel does not exist" })
		return
	}

	userContext, _ := c.Get("user")

	user := userContext.(UserSchema)
	userIDString := user.ID.String()[10:34]

	fmt.Println(userIDString)

	if _, ok := foundChannel.PrivateKeys[userIDString]; ok {
		_, _ = Channels.UpdateOne(context.TODO(), bson.D{{"_id", convertID}}, bson.D{{"$unset", bson.D{{"privatekeys." + userIDString, ""}}}})

		c.JSON(200, gin.H{"message": "You have left channel " + form.ChannelID})
		return
	}

	c.JSON(400, gin.H{"message": "You are not in this channel"})
}