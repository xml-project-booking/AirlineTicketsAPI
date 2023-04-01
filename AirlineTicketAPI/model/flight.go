package model

import (
	"encoding/json"
	"io"
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

type SearchCriteria struct {
	From         string `bson:"from" json:"from"`
	To           string `bson:"to" json:"to"`
	TicketNumber int    `bson:"number" json:"number"`
	Date         string `bson:"date" json:"date"`
}
type Flights []*Flight

func (u *Flights) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(u)
}

func (u *Flight) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(u)
}

func (u *Flight) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(u)
}
func (u *SearchCriteria) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(u)
}

func (u *SearchCriteria) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(u)
}
