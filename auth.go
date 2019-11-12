package main

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

var jwtKey = []byte("my_secret_key")

type RegisterStruct struct {
	Username string
	Password string
	PublicKey string
	ProtectedKey string
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

	if form.ProtectedKey == "" {
		c.JSON(400, gin.H{"message": "You must send a protected key"})
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

	ourUser := UserSchema{primitive.NewObjectID(), form.Username, hashedPassword, form.PublicKey, form.ProtectedKey}

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

type LoginStruct struct {
	Username string
	Password string
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

	c.JSON(200, gin.H{"user": user})
}

type Claims struct {
	User UserSchema
	jwt.StandardClaims
}

func EnsureAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Request.Header.Get("Authorization")

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