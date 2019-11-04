package main

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

var jwtKey = []byte("my_secret_key")

func RegisterRoute(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	publicKey := c.PostForm("publicKey")
	protectedKey := c.PostForm("protectedKey")

	if publicKey == "" {
		c.JSON(400, gin.H{"message": "You must send a public key"})
		return
	}

	if protectedKey == "" {
		c.JSON(400, gin.H{"message": "You must send a protected key"})
		return
	}

	var foundUser UserSchema

	err := Users.FindOne(context.TODO(), bson.D{{"username", username}}).Decode(&foundUser)

	if foundUser != (UserSchema{}) {
		c.JSON(400, gin.H{ "message": "Username already exists" })
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 13)

	if err != nil {
		c.JSON(400, gin.H{ "message": "Fatal error hashing password", "err": err })
		return
	}

	var hashedPassword = string(hash)

	ourUser := UserSchema{primitive.NewObjectID(), username, hashedPassword, publicKey, protectedKey}

	insertResult, err := Users.InsertOne(context.TODO(), ourUser)

	if err != nil {
		c.JSON(400, gin.H{ "message": "Error inserting user into db", "err": err })
		return
	}

	tokenString, err := createJWTTokenString(ourUser)

	if err != nil {
		c.JSON(500, gin.H{"message": "Error creating JWT token", "err": err})
		return
	}

	c.JSON(200, gin.H{"message": "Account created successfully", "insertedID": insertResult.InsertedID, "token": tokenString})
}

func LoginRoute(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	var foundUser UserSchema

	err := Users.FindOne(context.TODO(), bson.D{{"username", username}}).Decode(&foundUser)

	if err != nil {
		c.JSON(400, gin.H{ "message": "Username not found" })
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(password))

	if err != nil {
		c.JSON(400, gin.H{ "message": "Password is incorrect" })
		return
	}

	tokenString, err := createJWTTokenString(foundUser)

	if err != nil {
		c.JSON(500, gin.H{"message": "Error creating JWT token", "err": err})
		return
	}

	c.JSON(200, gin.H{"message": "Login successful", "token": tokenString})
}

func GetSessionRoute(c *gin.Context) {
	userContext, _ := c.Get("user")

	user := userContext.(UserSchema)

	c.JSON(200, gin.H{"user": user})
}

type Claims struct {
	User UserSchema
	jwt.StandardClaims
}

func EnsureAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("token")

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "No JWT token found"})
			return
		}

		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Your token is invalid"})
				return
			}

			fmt.Println(err)

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Bad request", "err": err})
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Your token is invalid"})
			return
		}

		claims.User.Password = "stop being snoopy"

		c.Set("user", claims.User)
		c.Set("token", token)

		c.Next()
	}
}

func createJWTTokenString(user UserSchema, ) (string, error) {
	user.PublicKey = ""
	user.ProtectedKey = ""
	claims := &Claims{
		User: user,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(5 * time.Hour * 24).Unix(),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtKey)
}