package model

import (
	"encoding/json"
	"io"
	"time"
	"golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Surname     string             `bson:"surname,omitempty" json:"surname"`
	PhoneNumber string             `bson:"phoneNumber,omitempty" json:"phoneNumber"`
	Email       string             `bson:"email" json:"email"`
	Username    string             `bson:"username,omitempty" json:"username"`
	Password    string             `bson:"password" json:"password"`
	BirthDate   time.Time          `bson:"birthdate,omitempty" json:"birthdate"`
	Role        Role               `bson:"role" json:"role"`
}

type Role int

const (
	Client = iota
	Admin
)

type Users []*User

func (u *Users) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(u)
}

func (u *User) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(u)
}

func (u *User) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(u)
}

func (user *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	user.Password = string(bytes)
	return nil
}

func (user *User) CheckPassword(providedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(providedPassword))
	if err != nil {
		return err
	}
	return nil
}
