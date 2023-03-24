package handlers

import (
	"Rest/model"
	"Rest/repo"
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
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

func (u *UserHandler) RegisterUser(rw http.ResponseWriter, h *http.Request) {

	user := h.Context().Value(KeyProduct{}).(*model.User)

	u.repo.Insert(user)

	rw.WriteHeader(http.StatusCreated)
}

func (u *UserHandler) UpdateUser(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]
	user := h.Context().Value(KeyProduct{}).(*model.User)

	u.repo.UpdateUser(id, user)
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

func (u *UserHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		u.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		next.ServeHTTP(rw, h)
	})
}
