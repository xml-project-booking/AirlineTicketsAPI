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

type KeyProduct1 struct{}
type FlightHandler struct {
	logger *log.Logger
	// NoSQL: injecting product repository
	repo *repo.FlightRepo
}

// Injecting the logger makes this code much more testable.
func NewFlightsHandler(l *log.Logger, r *repo.FlightRepo) *FlightHandler {
	return &FlightHandler{l, r}
}

func (u *FlightHandler) GetAllFlights(rw http.ResponseWriter, h *http.Request) {
	flights, err := u.repo.GetAll()
	if err != nil {
		u.logger.Print("Database exception: ", err)
	}

	if flights == nil {
		return
	}

	err = flights.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		u.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (u *FlightHandler) GetFlightById(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	flight, err := u.repo.GetById(id)
	if err != nil {
		u.logger.Print("Database exception: ", err)
	}

	if flight == nil {
		http.Error(rw, "Flight with given id not found", http.StatusNotFound)
		u.logger.Printf("Flight with id: '%s' not found", id)
		return
	}

	err = flight.ToJSON(rw)
	if err != nil {
		http.Error(rw, "Unable to convert to json", http.StatusInternalServerError)
		u.logger.Fatal("Unable to convert to json :", err)
		return
	}
}

func (u *FlightHandler) CreateFlight(rw http.ResponseWriter, h *http.Request) {
	flightDTO := h.Context().Value(KeyProduct{}).(*model.Flight)
	flight := model.Flight{To: flightDTO.To, From: flightDTO.From, Price: flightDTO.Price, FreeSeats: flightDTO.FreeSeats, Date: flightDTO.Date}
	u.repo.Insert(&flight)
	rw.WriteHeader(http.StatusCreated)
	json.NewEncoder(rw).Encode(flight)
	rw.Header().Set("Content-Type", "application/json")
}

func (p *FlightHandler) DeleteFlight(rw http.ResponseWriter, h *http.Request) {
	vars := mux.Vars(h)
	id := vars["id"]

	p.repo.Delete(id)
	rw.WriteHeader(http.StatusNoContent)
}

func (u *FlightHandler) MiddlewareFlightDeserialization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		user := &model.Flight{}
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

func (u *FlightHandler) MiddlewareAuthDeserialization(next http.Handler) http.Handler {
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

func (u *FlightHandler) MiddlewareContentTypeSet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, h *http.Request) {
		u.logger.Println("Method [", h.Method, "] - Hit path :", h.URL.Path)

		next.ServeHTTP(rw, h)
	})
}
