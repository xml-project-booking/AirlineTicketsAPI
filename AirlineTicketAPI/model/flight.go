package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Flight struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	From      string             `bson:"from" json:"from"`
	To        string             `bson:"to,omitempty" json:"to"`
	Price     float32            `bson:"price,omitempty" json:"price"`
	FreeSeats int                `bson:"freeseats" json:"freeseats"`
	Date      time.Time          `bson:"date,omitempty" json:"date"`
}

type Ticket struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserId   string             `bson:"userid,omitempty" json:"userid"`
	FlightId string             `bson:"flightid,omitempty" json:"flightid"`
}
