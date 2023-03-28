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
	//"github.com/rs/cors"
	gorillaHandlers "github.com/gorilla/handlers"
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
	storeLogger := log.New(os.Stdout, "[user-store] ", log.LstdFlags)

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

	//Registration
	registerUserRouter := router.Methods(http.MethodPost).Subrouter()
	registerUserRouter.HandleFunc("/registration", usersHandler.RegisterUser)
	registerUserRouter.Use(usersHandler.MiddlewareUserDeserialization)

	getByEmailRouter := router.Methods(http.MethodGet).Subrouter()
	getByEmailRouter.HandleFunc("/existsEmail/{email}", usersHandler.GetUserByEmail)

	getByUsernameRouter := router.Methods(http.MethodGet).Subrouter()
	getByUsernameRouter.HandleFunc("/existsUsername/{username}", usersHandler.GetUserByUsername)

	//Login
	loginUserRouter := router.Methods(http.MethodPost).Subrouter()
	loginUserRouter.HandleFunc("/login", usersHandler.LoginUser)
	loginUserRouter.Use(usersHandler.MiddlewareAuthDeserialization)
	//Proba autorizacije
	probaautRouter := router.Methods(http.MethodPost).Subrouter()
	probaautRouter.HandleFunc("/proba", usersHandler.ProbaAut)
	probaautRouter.Use(usersHandler.IsAuthorizedAdmin)

	//cors := gorillaHandlers.CORS(gorillaHandlers.AllowedMethods([]string{"*"}))

	headersOk := gorillaHandlers.AllowedHeaders([]string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization",
	"accept", "origin", "Cache-Control", "X-Requested-With"})
	originsOk := gorillaHandlers.AllowedOrigins([]string{"*"})
	methodsOk := gorillaHandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	cors := gorillaHandlers.CORS(headersOk, originsOk, methodsOk)
	//Initialize the server 
	server := http.Server{
		Addr:         ":" + port,
	 	Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
	 	ReadTimeout:  5 * time.Second,
	 	WriteTimeout: 5 * time.Second,
	}

	//Initialize the server
	// server := http.Server{
	// 	Addr:         ":" + port,
	// 	Handler:      cors.Default().Handler(router),
	// 	IdleTimeout:  120 * time.Second,
	// 	ReadTimeout:  1 * time.Second,
	// 	WriteTimeout: 1 * time.Second,
	// }

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
