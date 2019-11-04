package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FindLikeUsersStruct struct {
	LikeName string
}

func FindLikeUsers(c *gin.Context) {
	var form FindLikeUsersStruct

	c.BindJSON(&form)

	if form.LikeName == "" {
		c.JSON(400, gin.H{"message": "You must send a username"})
		return
	}

	cursor, err := Users.Find(context.TODO(), bson.D{{"username", primitive.Regex{Pattern: form.LikeName, Options: ""}}})

	users := make(map[string]UserSchema)

	if err != nil {
		c.JSON(500, gin.H{"message": "Could not find like users", "err": err})
		return
	}

	for cursor.Next(context.Background()) {

		user := UserSchema{}
		err := cursor.Decode(&user)
		if err != nil {
			//handle err
		}

		user.Password = ""
		user.ProtectedKey = ""
		user.PublicKey = ""

		users[user.Username] = user // you need to handle this in a for loop or something... I'm assuming there is only one result per id
	}

	c.JSON(200, gin.H{"message": "Successful", "results": users})
}