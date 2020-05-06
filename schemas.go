package main

import (
	"github.com/SherClockHolmes/webpush-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type UserSchema struct {
	ID primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	Username string `json:"username""`
	Password string `json:"password""`
	PublicKey string `json:"publicKey""`
}

type ChannelSchema struct {
	ID primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	Name string
	PrivateKeys map[string]string // [userID]: Channels symmetric AES key is encrypted with the select users public key
	UserMap map[string]string
}

type MessageSchema struct {
	ID primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	ChannelID primitive.ObjectID
	Timestamp time.Time
	Encrypted string
}

type SubscriptionSchema struct {
	ID primitive.ObjectID `bson:"_id" json:"_id,omitempty"`
	UserID primitive.ObjectID
	Type string
	Endpoint string
	ExpirationTime string
	Keys webpush.Keys
}