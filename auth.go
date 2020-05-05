package main

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var jwtKey = []byte("my_secret_key")

type RegisterStruct struct {
	Username string
	Password string
	PublicKey string
}

type LoginStruct struct {
	Username string
	Password string
}

func RegisterRoute(c *gin.Context) {
	var form RegisterStruct

	c.BindJSON(&form)

	if form.Username == "" {
		c.JSON(400, gin.H{"message": "You must send a username"})
		return
	}

	if form.Password == "" {
		c.JSON(400, gin.H{"message": "You must send a password"})
		return
	}

	if form.PublicKey == "" {
		c.JSON(400, gin.H{"message": "You must send a public key"})
		return
	}

	var foundUser UserSchema

	err := Users.FindOne(context.TODO(), bson.D{{"username", form.Username}}).Decode(&foundUser)

	if foundUser != (UserSchema{}) {
		c.JSON(400, gin.H{ "message": "Username already exists" })
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(form.Password), 17)

	if err != nil {
		c.JSON(400, gin.H{ "message": "Fatal error hashing password", "err": err })
		return
	}

	var hashedPassword = string(hash)

	ourUser := UserSchema{primitive.NewObjectID(), form.Username, hashedPassword, form.PublicKey}

	_, err = Users.InsertOne(context.TODO(), ourUser)

	if err != nil {
		c.JSON(400, gin.H{ "message": "Error inserting user into db", "err": err })
		return
	}

	tokenString, err := createJWTTokenString(ourUser)

	if err != nil {
		c.JSON(500, gin.H{"message": "Error creating JWT token", "err": err})
		return
	}

	c.JSON(200, gin.H{"message": "Account created successfully", "user": ourUser, "token": tokenString})
}

func LoginRoute(c *gin.Context) {
	var form LoginStruct

	c.BindJSON(&form)

	var foundUser UserSchema

	err := Users.FindOne(context.TODO(), bson.D{{"username", form.Username}}).Decode(&foundUser)

	if err != nil {
		c.JSON(400, gin.H{ "message": "Username not found" })
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(form.Password))

	if err != nil {
		c.JSON(400, gin.H{ "message": "Password is incorrect" })
		return
	}

	tokenString, err := createJWTTokenString(foundUser)

	if err != nil {
		c.JSON(500, gin.H{"message": "Error creating JWT token", "err": err})
		return
	}

	c.JSON(200, gin.H{"message": "Login successful", "token": tokenString, "user": foundUser})
}

func GetSessionRoute(c *gin.Context) {
	userContext, _ := c.Get("user")

	user := userContext.(UserSchema)

	var foundUser UserSchema

	err := Users.FindOne(context.TODO(), bson.D{{ "_id", user.ID }}).Decode(&foundUser)

	if err != nil {
		c.JSON(400, gin.H{"err": "User does not exist"})
		return
	}

	c.JSON(200, gin.H{"user": user, "userDB": foundUser})
}

func createJWTTokenString(user UserSchema, ) (string, error) {
	user.PublicKey = ""
	claims := &Claims{
		User: user,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(5 * time.Hour * 24).Unix(),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtKey)
}