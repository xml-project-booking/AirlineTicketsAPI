package handlers

import (
	"Rest/model"
	"Rest/repo"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type KeyProduct struct{}
type UserHandler struct {
	logger *log.Logger
	// NoSQL: injecting product repository
	repo *repo.UserRepo
}

// Injecting the logger makes this code much more testable.
func NewUsersHandler(l *log.Logger, r *repo.UserRepo) *UserHandler {
	return &UserHandler{l, r}
}

func (u *UserHandler) GetAllUsers(rw http.ResponseWriter, h *http.Request) {
	users, err := u.repo.GetAll()
	if err != nil {
		u.logger.Print("Database exception: ", err)
	}

	if users == nil {
		return
	}

	err = users.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		u.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (u *UserHandler) GetUserById(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	user, err := u.repo.GetById(id)
	if err != nil {
		u.logger.Print("Database exception: ", err)
	}

	if user == nil {
		http.Error(rw, "Patient with given id not found", http.StatusNotFound)
		u.logger.Printf("Patient with id: '%s' not found", id)
		return
	}

	err = user.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		u.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (u *UserHandler) GetUserByEmail(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	email := vars["email"]

	user, err := u.repo.GetByEmail(email)
	if err != nil {
		u.logger.Print("Database exception: ", err)
	}

	err = user.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		u.logger.Fatal("Unable to convert to json :", err)
		return
	}
}
func (u *UserHandler) GetUserByUsername(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	username := vars["username"]

	user, err := u.repo.GetByUsername(username)
	if err != nil {
		u.logger.Print("Database exception: ", err)
	}

	err = user.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		u.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (u *UserHandler) FindByEmail(email string) (*model.User, error) {
	user, err := u.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserHandler) FindByUsername(username string) (*model.User, error) {
	user, err := u.repo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserHandler) RegisterUser(rw http.ResponseWriter, h *http.Request) {
	userDTO := h.Context().Value(KeyProduct{}).(*model.User)
	hashPw, _ := HashPassword(userDTO.Password)
	user := model.User{Name: userDTO.Name, Surname: userDTO.Surname, PhoneNumber: userDTO.PhoneNumber, Email: userDTO.Email, Username: userDTO.Username, Password: hashPw, BirthDate: userDTO.BirthDate, Role: 0}

	existsEmail, _ := u.FindByEmail(user.Email)
	if existsEmail != nil {
		http.Error(rw, "User with given email already exists", http.StatusBadRequest)
		u.logger.Printf("User with email: '%s' already exists", existsEmail.Email)
		return
	}
	existsUsername, _ := u.FindByUsername(user.Username)
	if existsUsername != nil {
		http.Error(rw, "User with given username already exists", http.StatusBadRequest)
		u.logger.Printf("User with username: '%s' already exists", existsUsername.Username)
		return
	}

	u.repo.Insert(&user)
	rw.WriteHeader(http.StatusCreated)
	json.NewEncoder(rw).Encode(user)
	rw.Header().Set("Content-Type", "application/json")
}

func (u *UserHandler) LoginUser(rw http.ResponseWriter, h *http.Request) {
	authdetails := h.Context().Value(KeyProduct{}).(*model.Authentication)

	if authdetails.Username == "" {
		http.Error(rw, "Username or Password is incorrect", http.StatusBadRequest)
		u.logger.Printf("Username or Password is incorrect")
		return
	}
	user, _ := u.FindByUsername(authdetails.Username)
	if user == nil {
		http.Error(rw, "User doesnt exists", http.StatusBadRequest)
		u.logger.Printf("User with username: '%s' doesnt exists", authdetails.Username)
		return
	}
	check := CheckPasswordHash(authdetails.Password, user.Password)

	if !check {
		http.Error(rw, "Username or Password is incorrect", http.StatusBadRequest)
		u.logger.Printf("Username or Password is incorrect")
		return
	}
	stringRole := "USER"
	if user.Role == 1 {
		stringRole = "ADMIN"
	}

	validToken, err := GenerateJWT(user.Email, stringRole)
	if err != nil {
		http.Error(rw, "Failed to genetare token", http.StatusBadRequest)
		u.logger.Printf("Failed to genetare token")
		return
	}

	var token model.Token
	token.Email = user.Email
	token.Role = stringRole
	token.TokenString = validToken
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(token)
}

func (u *UserHandler) UpdateUser(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]
	user := h.Context().Value(KeyProduct{}).(*model.User)

	u.repo.UpdateUser(id, user)
	rw.WriteHeader(http.StatusOK)
}

func (u *UserHandler) ProbaAut(rw http.ResponseWriter, h *http.Request) {
	if h.Header["Role"][0] != "ADMIN" {
		http.Error(rw, "You're not admin", http.StatusUnauthorized)
		return
	}
	u.logger.Printf("Admin prijavljen")
	rw.WriteHeader(http.StatusOK)
}

func (u *UserHandler) MiddlewareUserDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		user := &model.User{}
		err := user.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			u.logger.Fatal(err)
			return
		}

		ctx := context.WithValue(h.Context(), KeyProduct{}, user)
		h = h.WithContext(ctx)

		next.ServeHTTP(rw, h)
	})
}
func (u *UserHandler) MiddlewareAuthDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		auth := &model.Authentication{}
		err := auth.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			u.logger.Fatal(err)
			return
		}

		ctx := context.WithValue(h.Context(), KeyProduct{}, auth)
		h = h.WithContext(ctx)

		next.ServeHTTP(rw, h)
	})
}

func (u *UserHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		u.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		next.ServeHTTP(rw, h)
	})
}

func (u *UserHandler) IsAuthorizedAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		tokenString := GetJWT(h.Header)
		var mySigningKey = []byte("secretkey")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("There was an error in parsing")
			}
			return mySigningKey, nil
		})

		if err != nil {
			http.Error(rw, "Your Token has been expired", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if claims["role"] == "ADMIN" {
				h.Header.Set("Role", "ADMIN")
				next.ServeHTTP(rw, h)
				return
			}
		}
		http.Error(rw, "Not Authorized", http.StatusUnauthorized)
	})
}

func (u *UserHandler) IsAuthorizedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		tokenString := GetJWT(h.Header)
		var mySigningKey = []byte("secretkey")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("there was an error in parsing")
			}
			return mySigningKey, nil
		})

		if err != nil {
			http.Error(rw, "Your Token has been expired", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if claims["role"] == "USER" {
				h.Header.Set("Role", "USER")
				next.ServeHTTP(rw, h)
				return
			}
		}
		http.Error(rw, "Not Authorized", http.StatusUnauthorized)
	})
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func GenerateJWT(email, role string) (string, error) {
	var mySigningKey = []byte("secretkey")
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["email"] = email
	claims["role"] = role
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, _ := token.SignedString(mySigningKey)

	return tokenString, nil
}
func GetJWT(r http.Header) string {
	bearToken := r.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}
