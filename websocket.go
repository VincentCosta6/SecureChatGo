package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func WebSocketUpgrade(c *gin.Context) {
	token := c.GetHeader("Authorization")

	if token == "" {
		token = c.GetHeader("Sec-Websocket-Protocol")[5:]
	}

	fmt.Println(c.Request.Header)

	fmt.Println(token)

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
}