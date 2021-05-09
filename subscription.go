package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"
	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
)

type SubscribeStruct struct {
	Endpoint string
	ExpirationTime string
	Type string
	Keys webpush.Keys
}

func SubscribeRoute(c *gin.Context) {
	var form SubscribeStruct

	c.BindJSON(&form)

	if form.Endpoint == "" {
		c.JSON(400, gin.H{"message": "You must send an endpoint URL"})
		return
	}

	var foundSubscription SubscriptionSchema

	err := Subscription.FindOne(context.TODO(), bson.D{{"endpoint", form.Endpoint}}).Decode(&foundSubscription)

	if err != nil {
		fmt.Println("error3")
		fmt.Println(err)
	}

	if foundSubscription != (SubscriptionSchema{}) {
		c.JSON(200, gin.H{ "message": "success", "subscription": foundSubscription, "active": true, "newlyCreated": false })
		return
	}

	userContext, _ := c.Get("user")

	user := userContext.(UserSchema)

	ourSubcription := SubscriptionSchema{primitive.NewObjectID(), user.ID, form.Type, form.Endpoint, form.ExpirationTime, form.Keys}

	_, err = Subscription.InsertOne(context.TODO(), ourSubcription)

	if err != nil {
		c.JSON(400, gin.H{ "message": "Error inserting subscription into db", "err": err })
		return
	}

	c.JSON(200, gin.H{"message": "Subscription created successfully", "subscription": ourSubcription, "active": true, "newlyCreated": true})
}

func SendNotif(key string, newMessage MessageSchema) {
	id, err := primitive.ObjectIDFromHex(key)

	if err != nil {
		fmt.Println("error5")
		fmt.Println(err)
	}

	cursor, err := Subscription.Find(context.TODO(), bson.D{{"userid", id}})

	if err != nil {
		fmt.Println("error4")
	}

	for cursor.Next(context.Background()) {
		subscription := SubscriptionSchema{}
		err := cursor.Decode(&subscription)

		if err != nil {
			fmt.Println("err", err)
			continue
		}

		if subscription != (SubscriptionSchema{}) {
			sendMessageWithSubscription(subscription, newMessage)
		}
	}
}

func sendMessageWithSubscription(subscription SubscriptionSchema, newMessage MessageSchema) {
	marshaled, _ := json.Marshal(newMessage)

	if subscription.Type == "expo" {
		go sendExpoNotif(subscription, marshaled)
		return
	} else {
		go sendWebpushNotif(subscription, marshaled)
	}
}

func sendExpoNotif(subscription SubscriptionSchema, marshaled []byte) {
	rpushToken, err := expo.NewExponentPushToken(subscription.Endpoint)
	if err != nil {
		panic(err)
		return
	}

	rpushArr := []expo.ExponentPushToken{ rpushToken }

	client := expo.NewPushClient(nil)

	response, err := client.Publish(
		&expo.PushMessage{
			To: rpushArr,
			Body: "You received a new message",
			Data: map[string]string{"message": string(marshaled)},
			Sound: "default",
			Title: "SecureChat",
			Priority: expo.DefaultPriority,
			TTLSeconds: 50000,
		},
	)

	if err != nil {
		panic(err)
		return
	}

	if response.ValidateResponse() != nil {
		fmt.Println(response.PushMessage.To, "failed")
	}
}

func sendWebpushNotif(subscription SubscriptionSchema, marshaled []byte) {
	s := &webpush.Subscription{subscription.Endpoint, subscription.Keys}

	res, err := webpush.SendNotification(marshaled, s, &webpush.Options{
		TTL:             50000,
		VAPIDPublicKey:  os.Getenv("PushPublicKey"),
		VAPIDPrivateKey: os.Getenv("PushPrivateKey"),
	})

	if err != nil {
		fmt.Println(err)
	}

	res.Body.Close()
}