package model

import (
	"encoding/json"
	"io"
)

type Authentication struct {
	Username    string `json:"username"`
	Password string `json:"password"`
}

type Token struct {
	Role        string `json:"role"`
	Email       string `json:"email"`
	TokenString string `json:"token"`
}

func (a *Authentication) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(a)
}

func (a *Authentication) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(a)
}