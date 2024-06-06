package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Account struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name"`
	Balance  float64            `json:"balance" bson:"balance"`
	Currency string             `json:"currency" bson:"currency"`
	Security *Secret            `json:"security" bson:"security"`
}

type Secret struct {
	NickName string `json:"nickname" bson:"nickname"`
	Password string `json:"password" bson:"password"`
}

type Payment struct {
	SenderID   primitive.ObjectID `json:"sender_id" bson:"sender_id"`
	ReceiverID primitive.ObjectID `json:"receiver_id" bson:"receiver_id"`
	Amount     float64            `json:"amount" bson:"amount"`
}
