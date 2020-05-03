package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Claims struct {
	User UserSchema
	jwt.StandardClaims
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

			fmt.Println("got it broger")
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