package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

var (
	Messages *mongo.Collection
	Users    *mongo.Collection
	Channels *mongo.Collection
	ChannelUsers *mongo.Collection
)


func main() {
	setupDB()
	initRouter()
}

func setupDB() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	Messages = client.Database("GoTest").Collection("messages")
	Users = client.Database("GoTest").Collection("users")
	Channels = client.Database("GoTest").Collection("channels")
	ChannelUsers = client.Database("GoTest").Collection("channelusers")
}

func initRouter() {
	r := gin.Default()

	r.POST("/register", RegisterRoute)
	r.POST("/login", LoginRoute)

	r.GET("/like/users", FindLikeUsers)

	r.GET("/get/session", EnsureAuth(), GetSessionRoute)
	r.POST("/channel/create", EnsureAuth(), CreateChannelRoute)
	r.POST("/message/create", EnsureAuth(), CreateMessageRoute)

	r.Run()
}
