package main

import (
	"Rest/handlers"
	"Rest/repo"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/cors"

	"github.com/gorilla/mux"
)

func main() {
	//Reading from environment, if not set we will default it to 8080.
	//This allows flexibility in different environments (for eg. when running multiple docker api's and want to override the default port)
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	// Initialize context
	timeoutContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//Initialize the logger we are going to use, with prefix and datetime for every log
	logger := log.New(os.Stdout, "[product-api] ", log.LstdFlags)
	storeLogger := log.New(os.Stdout, "[patient-store] ", log.LstdFlags)

	// NoSQL: Initialize Product Repository store

	storeUser, err := repo.NewUserRepo(timeoutContext, storeLogger)
	if err != nil {
		logger.Fatal(err)
	}
	defer storeUser.DisconnectUserRepo(timeoutContext)

	// NoSQL: Checking if the connection was established
	storeUser.PingUserRepo()

	//Initialize the handler and inject said logger

	usersHandler := handlers.NewUsersHandler(logger, storeUser)

	//Initialize the router and add a middleware for all the requests
	router := mux.NewRouter()

	router.Use(usersHandler.MiddlewareContentTypeSet)

	registerUserRouter := router.Methods(http.MethodPost).Subrouter()
	registerUserRouter.HandleFunc("/registration", usersHandler.RegisterUser)
	registerUserRouter.Use(usersHandler.MiddlewareUserDeserialization)

	//cors := gorillaHandlers.CORS(gorillaHandlers.AllowedMethods([]string{"*"}))

	//Initialize the server
	server := http.Server{
		Addr:         ":" + port,
		Handler:      cors.Default().Handler(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	logger.Println("Server listening on port", port)
	//Distribute all the connections to goroutines
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Fatal(err)
		}
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	signal.Notify(sigCh, os.Kill)

	sig := <-sigCh
	logger.Println("Received terminate, graceful shutdown", sig)

	//Try to shutdown gracefully
	if server.Shutdown(timeoutContext) != nil {
		logger.Fatal("Cannot gracefully shutdown...")
	}
	logger.Println("Server stopped")
}
