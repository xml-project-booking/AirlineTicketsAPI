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
type UserRepo struct {
	cli    *mongo.Client
	logger *log.Logger
}

// NoSQL: Constructor which reads db configuration from environment
func NewUserRepo(ctx context.Context, logger *log.Logger) (*UserRepo, error) {
	dburi := os.Getenv("MONGO_DB_URI")

	client, err := mongo.NewClient(options.Client().ApplyURI(dburi))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &UserRepo{
		cli:    client,
		logger: logger,
	}, nil
}

// Disconnect from database
func (ur *UserRepo) DisconnectUserRepo(ctx context.Context) error {
	err := ur.cli.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Check database connection
func (ur *UserRepo) PingUserRepo() {
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

func (ur *UserRepo) GetAll() (model.Users, error) {
	// Initialise context (after 5 seconds timeout, abort operation)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := ur.getCollection()

	var users model.Users
	usersCursor, err := usersCollection.Find(ctx, bson.M{})
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	if err = usersCursor.All(ctx, &users); err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	return users, nil
}

func (ur *UserRepo) GetById(id string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCollection := ur.getCollection()

	var user model.User
	objID, _ := primitive.ObjectIDFromHex(id)
	err := usersCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		ur.logger.Println(err)
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepo) Insert(user *model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	usersCollection := ur.getCollection()

	result, err := usersCollection.InsertOne(ctx, &user)
	if err != nil {
		ur.logger.Println(err)
		return err
	}
	ur.logger.Printf("Documents ID: %v\n", result.InsertedID)
	return nil
}

func (ur *UserRepo) UpdateUser(id string, user *model.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	usersCollection := ur.getCollection()

	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{
		"name":    user.Name,
		"surname": user.Surname,
	}}
	result, err := usersCollection.UpdateOne(ctx, filter, update)
	ur.logger.Printf("Documents matched: %v\n", result.MatchedCount)
	ur.logger.Printf("Documents updated: %v\n", result.ModifiedCount)

	if err != nil {
		ur.logger.Println(err)
		return err
	}
	return nil
}

func (ur *UserRepo) getCollection() *mongo.Collection {
	userDatabase := ur.cli.Database("mongoDemo")
	usersCollection := userDatabase.Collection("users")
	return usersCollection
}
