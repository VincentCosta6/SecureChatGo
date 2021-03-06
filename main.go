package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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
	Subscription *mongo.Collection
)

func main() {
	/*if InitializeKeys() == false {
		log.Fatal("[Web Push] Fatal error occurred during initialization of keys")
	}*/

	setupDB()
	initializeRouter()
}

func setupDB() {
	if os.Getenv("local") == "true" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
	credential := options.Credential{
		Username: os.Getenv("MONGO_USERNAME"),
		Password: os.Getenv("MONGO_PASSWORD"),
	}
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_CONNECTION")).SetAuth(credential)

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
	Subscription = client.Database("GoTest").Collection("subscriptions")
}

var HubGlob = NewHub()

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
	port := os.Getenv("PORT")

	r.POST("/register", RegisterRoute)
	r.POST("/login", LoginRoute)

	r.GET("/like/users/:LikeName", FindLikeUsers)

	r.GET("/get/session", EnsureAuth(), GetSessionRoute)

	r.POST("/channel/create", EnsureAuth(), CreateChannelRoute)
	r.POST("/channel/add/user", EnsureAuth(), AddUserRoute)
	r.GET("/channels/mine", EnsureAuth(), FindChannels)
	r.DELETE("/channel/leave", EnsureAuth(), LeaveChannelRoute)

	r.POST("/message/create", EnsureAuth(), CreateMessageRoute)
	r.GET("/channels/messages/:channelID", EnsureAuth(), GetMessagesRoute)
	r.POST("/upload", EnsureAuth(), FileUploadRoute)
	r.Static("/download-file", "./files")

	r.POST("/subscription", EnsureAuth(), SubscribeRoute)

	// r.RunTLS(":443", "./server.pem", "./server.key")
	r.Run(":" + port)
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
func initializePublicPrivateKeys() {

}