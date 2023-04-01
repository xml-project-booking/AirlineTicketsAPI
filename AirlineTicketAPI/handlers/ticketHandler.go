package handlers

import (
	"Rest/model"
	"Rest/repo"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type TicketHandler struct {
	logger *log.Logger
	// NoSQL: injecting product repository
	repo *repo.TicketRepo
}

// Injecting the logger makes this code much more testable.
func NewTicketsHandler(l *log.Logger, r *repo.TicketRepo) *TicketHandler {
	return &TicketHandler{l, r}
}

func (u *TicketHandler) GetTicketById(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	ticket, err := u.repo.GetById(id)
	if err != nil {
		u.logger.Print("Database exception: ", err)
	}

	if ticket == nil {
		http.Error(rw, "Patient with given id not found", http.StatusNotFound)
		u.logger.Printf("Patient with id: '%s' not found", id)
		return
	}

	err = ticket.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		u.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (u *TicketHandler) CreateTicket(rw http.ResponseWriter, h *http.Request) {
	ticketDTO := h.Context().Value(KeyProduct{}).(*model.Ticket)
	ticket := model.Ticket{FlightId: ticketDTO.FlightId, UserId: ticketDTO.UserId, NumberOfSeats: ticketDTO.NumberOfSeats}
	u.repo.Insert(&ticket)
	rw.WriteHeader(http.StatusCreated)
	json.NewEncoder(rw).Encode(ticket)
	rw.Header().Set("Content-Type", "application/json")
}

func (u *TicketHandler) MiddlewareTicketDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		user := &model.Ticket{}
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

func (u *TicketHandler) MiddlewareAuthDeserialization(next http.Handler) http.Handler {
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

func (u *TicketHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		u.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		next.ServeHTTP(rw, h)
	})
}
