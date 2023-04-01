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
type TicketRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

// NoSQL: Constructor which reads db configuration from environment
func NewTicketRepo(ctx context.Context, logger *log.Logger) (*TicketRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &TicketRepo{
		cli:    client,
		logger: logger,
	}, nil
}

// Disconnect from database
func (ur *TicketRepo) DisconnectTicketRepo(ctx context.Context) error {
	err := ur.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Check database connection
func (ur *TicketRepo) PingTicketRepo() {
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

func (ur *TicketRepo) GetById(id string) (*model.Ticket, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ticketCollection := ur.getCollection()

	var ticket model.Ticket
	objID, _ := primitive.ObjectIDFromHex(id)
	err := ticketCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&ticket)
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	return &ticket, nil
}

func (ur *TicketRepo) Insert(ticket *model.Ticket) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ticketsCollection := ur.getCollection()

	result, err := ticketsCollection.InsertOne(ctx, &ticket)
	if err != nil {
		ur.logger.Println(err)
		return err
	}
	ur.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}

func (ur *TicketRepo) getCollection() *mongo.Collection {
	ticketDatabase := ur.cli.Database("mongoDemo")
	ticketsCollection := ticketDatabase.Collection("tickets")
	return ticketsCollection
}
