package model

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
)

type Ticket struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserId        string             `bson:"userId" json:"userId"`
	FlightId      string             `bson:"flightId" json:"flightId"`
	NumberOfSeats int                `bson:"numberOfSeats" json:"numberOfSeats"`
}

type Tickets []*Ticket

func (u *Tickets) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(u)
}
