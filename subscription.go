package main

import (
	"context"
	"fmt"
	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	fmt.Println(form.Keys.Auth)
	fmt.Println(form.Keys.P256dh)

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
	return
}