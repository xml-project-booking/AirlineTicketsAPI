package repo

import (
	"Rest/model"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	// NoSQL: module containing Mongo api client
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// NoSQL: ProductRepo struct encapsulating Mongo api client
type FlightRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

// NoSQL: Constructor which reads db configuration from environment
func NewFlightRepo(ctx context.Context, logger *log.Logger) (*FlightRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &FlightRepo{
		cli:    client,
		logger: logger,
	}, nil
}

// Disconnect from database
func (ur *FlightRepo) DisconnectFlightRepo(ctx context.Context) error {
	err := ur.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Check database connection
func (ur *FlightRepo) PingFlightRepo() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := ur.cli.Ping(ctx, readpref.Primary())
	if err != nil {
		ur.logger.Println(err)
	}

	// Print available databases
	databases, err := ur.cli.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		ur.logger.Println(err)
	}
	fmt.Println(databases)
}

func (ur *FlightRepo) GetAll() (model.Flights, error) {
	// Initialise context (after 5 seconds timeout, abort operation)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	flightsCollection := ur.getCollection()

	var flights model.Flights
	usersCursor, err := flightsCollection.Find(ctx, bson.M{})
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	if err = usersCursor.All(ctx, &flights); err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	return flights, nil
}
func (pr *FlightRepo) GetBySearchCriteria(search *model.SearchCriteria) (model.Flights, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	flightsCollection := pr.getCollection()
	date, err := time.Parse(time.RFC3339, search.Date)
	fromDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	toDate := time.Date(date.Year(), date.Month(), date.Day()+1, 0, 0, 0, 0, time.UTC)
	var flights model.Flights
	filter := bson.D{
		{Key: "$and",
			Value: bson.A{
				bson.D{{Key: "to", Value: bson.D{{Key: "$regex", Value: search.To}}}},
				bson.D{{Key: "from", Value: bson.D{{Key: "$regex", Value: search.From}}}},
				bson.D{{Key: "freeseats", Value: bson.D{{Key: "$gt", Value: search.TicketNumber}}}},
				bson.D{{Key: "date", Value: bson.M{
					"$gt": fromDate,
					"$lt": toDate,
				}}},
			},
		},
	}

	patientsCursor, err := flightsCollection.Find(ctx, filter)

	if err != nil {
		pr.logger.Println(err)
		return nil, err
	}
	if err = patientsCursor.All(ctx, &flights); err != nil {
		pr.logger.Println(err)
		return nil, err
	}

	return flights, nil
}

func (ur *FlightRepo) GetById(id string) (*model.Flight, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := ur.getCollection()

	var flight model.Flight
	objID, _ := primitive.ObjectIDFromHex(id)
	err := usersCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&flight)
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	return &flight, nil
}

func (ur *FlightRepo) Insert(flight *model.Flight) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	flightsCollection := ur.getCollection()

	result, err := flightsCollection.InsertOne(ctx, &flight)
	if err != nil {
		ur.logger.Println(err)
		return err
	}
	ur.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}

func (ur *FlightRepo) UpdateFlight(id string, flight *model.Flight) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	flightCollection := ur.getCollection()

	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{
		"freeseats": flight.FreeSeats,
		"price":     flight.Price,
	}}
	result, err := flightCollection.UpdateOne(ctx, filter, update)
	ur.logger.Printf("Documents matched: %v\n", result.MatchedCount)
	ur.logger.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		ur.logger.Println(err)
		return err
	}
	return nil
}
func (pr *FlightRepo) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	patientsCollection := pr.getCollection()

	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "_id", Value: objID}}
	result, err := patientsCollection.DeleteOne(ctx, filter)
	if err != nil {
		pr.logger.Println(err)
		return err
	}
	pr.logger.Printf("Documents deleted: %v\n", result.DeletedCount)
	return nil
}

func (ur *FlightRepo) getCollection() *mongo.Collection {
	userDatabase := ur.cli.Database("mongoDemo")
	flightsCollection := userDatabase.Collection("flights")
	return flightsCollection
}
