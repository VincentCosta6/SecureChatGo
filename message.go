package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type CreateMessageType struct {
	ChannelID string
	Message string
}

type AddUser struct {
	ChannelID primitive.ObjectID
	NewUsers map[string]string
	UserMap map[string]string
}

type WebsocketMessageType struct {
	MessageType string
	MessageContent interface{}
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

	err = Channels.FindOne(context.TODO(), bson.D{{"_id", convertID}}).Decode(&foundChannel)

	if err != nil {
		c.JSON(400, gin.H{ "message": "Channel does not exist" })
		return
	}

	newMessage := MessageSchema{primitive.NewObjectID(), convertID, time.Now(), form.Message}

	go (func() {
		message := WebsocketMessageType{"NEW_MESSAGE", newMessage}

		clients := make([]string, 0)

		for key := range foundChannel.PrivateKeys {
			clients = append(clients, key)
			go sendNotif(key, newMessage)
		}

		HubGlob.createMessage <- CreatedMessageStruct{message:&message, clients: &clients}
	})()

	_, err = Messages.InsertOne(context.TODO(), newMessage)

	if err != nil {
		c.JSON(400, gin.H{"message": "Could not insert message", "err": err})
		return
	}

	c.JSON(200, gin.H{"message": "Message sent successfully", "inserted": newMessage})
}

func sendNotif(key string, newMessage MessageSchema) {
	var foundSubscription SubscriptionSchema

	id, err := primitive.ObjectIDFromHex(key)

	if err != nil {
		fmt.Println("error5")
		fmt.Println(err)
		fmt.Println(foundSubscription)
	}

	err = Subscription.FindOne(context.TODO(), bson.D{{"userid", id}}).Decode(&foundSubscription)

	if err != nil {
		fmt.Println("error4")
		fmt.Println(err)
		fmt.Println(foundSubscription)
	}

	if foundSubscription != (SubscriptionSchema{}) {
		s := &webpush.Subscription{foundSubscription.Endpoint, foundSubscription.Keys}

		marshaled, err := json.Marshal(newMessage)

		res, err := webpush.SendNotification(marshaled, s, &webpush.Options{
			TTL:             50000,
			VAPIDPublicKey:  PushPublicKey,
			VAPIDPrivateKey: PushPrivateKey,
		})

		if err != nil {
			fmt.Println(err)
		}

		res.Body.Close()
	}
}

func GetMessagesRoute(c * gin.Context) {
	idQuery := c.Param("channelID")

	if idQuery == "" {
		c.JSON(400, gin.H{"message": "You must send a channelID in the path param"})
		return
	}

	id, err := primitive.ObjectIDFromHex(idQuery)

	if err != nil {
		c.JSON(400, gin.H{"message": "Could not parse id into an ObjectID", "err": err})
		return
	}

	messages := make([]MessageSchema, 0)

	cursor, err := Messages.Find(context.TODO(), bson.M{"channelid": id})

	if err != nil {
		c.JSON(500, gin.H{"message": "Could not query messages", "err": err})
		return
	}

	for cursor.Next(context.Background()) {

		message := MessageSchema{}
		err := cursor.Decode(&message)
		if err != nil {
			//handle err
			fmt.Println("err", err)
		}

		messages = append(messages, message)
	}

	c.JSON(200, gin.H{"message": "Success", "messages": messages})
}