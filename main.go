package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var (
	Messages *mongo.Collection
	Users    *mongo.Collection
	Channels *mongo.Collection
	ChannelUsers *mongo.Collection
	FileMetaData *mongo.Collection
)

func main() {
	setupDB()
	initializeRouter()
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
	FileMetaData = client.Database("GoTest").Collection("filedata")
}

var HubGlob = newHub()

func initializeMiscRoutes(r *gin.Engine) {
	pingStatus := bson.M{"msg": "healthy"}

	r.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "./website.html")
	})

	r.GET("/ping", func (c *gin.Context) {
		c.JSON(200, pingStatus)
	})

	r.GET("/download/windows/SecureChat-Setup-0.5.8.exe", ServeExecutable)
}

func initializeWebSocketRoute(r *gin.Engine) {
	r.GET("/ws", WebSocketUpgrade)
}

func initializeMainRoutes(r *gin.Engine) {
	r.POST("/register", RegisterRoute)
	r.POST("/login", LoginRoute)

	r.GET("/like/users/:LikeName", FindLikeUsers)

	r.GET("/get/session", EnsureAuth(), GetSessionRoute)

	r.POST("/channel/create", EnsureAuth(), CreateChannelRoute)
	r.POST("/channel/add/user", EnsureAuth(), AddUserRoute)
	r.GET("/channels/mine", EnsureAuth(), FindChannels)

	r.POST("/message/create", EnsureAuth(), CreateMessageRoute)
	r.GET("/channels/messages/:channelID", EnsureAuth(), GetMessagesRoute)
	r.POST("/upload", EnsureAuth(), FileUploadRoute)
	r.Static("/download-file", "./files")

	r.RunTLS(":443", "./server.pem", "./server.key")
}

func initializeRouter() {
	//gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(CORSMiddleware())

	go HubGlob.run()

	initializeMiscRoutes(r)
	initializeWebSocketRoute(r)
	initializeMainRoutes(r)

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan struct{})

	fmt.Println("Waiting for interrupt")

	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		fmt.Println("\nReceived an interrupt, stopping services...\n")
		close(cleanupDone)
	}()
	<-cleanupDone
}