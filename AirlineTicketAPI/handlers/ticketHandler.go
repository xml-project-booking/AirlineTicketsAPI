package handlers

import (
	"Rest/model"
	"Rest/repo"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type TicketHandler struct {
	logger *log.Logger
	// NoSQL: injecting product repository

	repo       *repo.TicketRepo
	flightRepo *repo.FlightRepo
	userRepo   *repo.UserRepo
}

// Injecting the logger makes this code much more testable.
func NewTicketsHandler(l *log.Logger, r *repo.TicketRepo, f *repo.FlightRepo, u *repo.UserRepo) *TicketHandler {
	return &TicketHandler{l, r, f, u}
}

func (u *TicketHandler) GetAllTicketsByUserId(rw http.ResponseWriter, h *http.Request) {

	ticketDTO := h.Context().Value(KeyProduct{}).(*model.Ticket)
	user, err := u.userRepo.GetByUsername(ticketDTO.UserId)
	tickets, err := u.repo.GetAllByUserId(user.ID.Hex())
	if err != nil {
		u.logger.Print("Database exception: ", err)
	}

	if tickets == nil {
		return
	}

	err = tickets.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		u.logger.Fatal("Unable to convert to json :", err)
		return
	}
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
	log.Println("H: ")
	ticketDTO := h.Context().Value(KeyProduct{}).(*model.Ticket)
	user, err := u.userRepo.GetByUsername(ticketDTO.UserId)
	if err != nil {
		log.Fatalf("An error occurred while fetching the user by username: %v", err)
	}

	ticket := model.Ticket{FlightId: ticketDTO.FlightId, UserId: user.ID.Hex(), NumberOfSeats: ticketDTO.NumberOfSeats}
	log.Println("FLIGHTID: " + ticket.FlightId + " | " + ticketDTO.FlightId)
	flight, err := u.flightRepo.GetById(ticketDTO.FlightId)
	if err != nil {
		log.Fatalf("An error occurred while fetching the flight: %v", err)
	}

	if flight.Date.Before(time.Now()) {
		log.Println("That flight has already departed")
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	if flight.FreeSeats < ticket.NumberOfSeats {
		log.Println("That flight doesn't have enough available seats")
		rw.WriteHeader(http.StatusNotAcceptable)
		return
	}

	flight.FreeSeats -= ticket.NumberOfSeats
	if err := u.flightRepo.UpdateFlight(ticketDTO.FlightId, flight); err != nil {
		log.Fatalf("An error occurred while updating the flight: %v", err)
	}

	if err := u.repo.Insert(&ticket); err != nil {
		log.Fatalf("An error occurred while inserting the ticket: %v", err)
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(rw).Encode(ticket); err != nil {
		log.Fatalf("An error occurred while encoding the response: %v", err)
	}
}

func (u *TicketHandler) MiddlewareTicketDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		ticket := &model.Ticket{}
		err := ticket.FromJSON(h.Body)
		if err != nil {
			http.Error(rw, "Unable to decode json", http.StatusBadRequest)
			u.logger.Fatal(err)
			return
		}

		ctx := context.WithValue(h.Context(), KeyProduct{}, ticket)
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
