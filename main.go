package main

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
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

var HubGlob = newHub()

func initRouter() {
	//gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(CORSMiddleware())

	go HubGlob.run()

	pingStatus := bson.M{"msg": "healthy"}

	r.GET("/ping", func (c *gin.Context) {
		c.JSON(200, pingStatus)
	})

	r.GET("/download/windows/SecureChat-Setup-0.5.2.exe", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "./SecureChat Setup 0.5.2.exe")
	})

	r.GET("/ws", func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		claims := &Claims{}

		tokenJWT, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				fmt.Println("Token invalid")
				return
			}

			fmt.Println("Bad request")
			return
		}

		if !tokenJWT.Valid {
			fmt.Println("Bad token")
			return
		}

		serveWs(HubGlob, c.Writer, c.Request, claims)
	})

	r.POST("/register", RegisterRoute)
	r.POST("/login", LoginRoute)

	r.GET("/like/users/:LikeName", FindLikeUsers)

	r.GET("/get/session", EnsureAuth(), GetSessionRoute)

	r.POST("/channel/create", EnsureAuth(), CreateChannelRoute)
	r.GET("/channels/mine", EnsureAuth(), FindChannels)

	r.POST("/message/create", EnsureAuth(), CreateMessageRoute)
	r.GET("/channels/messages/:channelID", EnsureAuth(), GetMessagesRoute)

	r.RunTLS(":443", "./server.pem", "./server.key")

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

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}