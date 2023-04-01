package model

import (
	"encoding/json"
	"io"
	"net/http"
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

func (t *Ticket) ToJSON(rw http.ResponseWriter) error {
	e := json.NewEncoder(rw)
	return e.Encode(t)
}

func (t *Ticket) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(t)
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
