package main

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type UserSchema struct {
	ID primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	Username string
	Password string
	PublicKey string
	ProtectedKey string // This is the users private key but it is encrypted with their password using AES
}

type ChannelSchema struct {
	ID primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	Name string
	PrivateKeys map[string]string // [userID]: Channels symmetric AES key is encrypted with the select users public key
}

type MessageSchema struct {
	ID primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	ChannelID primitive.ObjectID
	Timestamp time.Time
	Encrypted string
}