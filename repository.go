package mrepo

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/uuid"
)

type mrepo struct {
	client        *mongo.Client
	isInitialized bool

	dbName string
}

type Repo interface {
	Init() error
	Done()
	IsInitialized() bool

	Insert(document Document) (Document, error)
	Update(document Document) (Document, error)
	FindByID(document Document) (Document, error)
}

func (mrepo *mrepo) Insert(document Document) (Document, error) {

	collection := mrepo.client.Database(mrepo.dbName).Collection(document.Collection())

	now := time.Now()
	toInsert := D(document.Collection(), document.Data(), map[string]interface{}{
		document.IDKey(): uuid.New().String(),
		"created_at":     now,
		"updated_at":     now,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, toInsert.Data())
	if err != nil {
		return nil, err
	}

	return toInsert, nil
}

func (mrepo *mrepo) Update(document Document) (Document, error) {

	collection := mrepo.client.Database(mrepo.dbName).Collection(document.Collection())

	now := time.Now()

	filter := bson.M{document.IDKey(): bson.M{"$eq": document.ID()}}

	toUpdate := D(document.Collection(), document.Data(), map[string]interface{}{
		"updated_at": now,
	})

	update := bson.M{
		"$set": bson.M(toUpdate.Data()),
	}

	_, err := collection.UpdateOne(
		context.Background(),
		filter,
		update,
	)

	if err != nil {
		return nil, err
	}

	return toUpdate, nil

}

func (mrepo *mrepo) FindByID(document Document) (Document, error) {

	collection := mrepo.client.Database(mrepo.dbName).Collection(document.Collection())

	filter := bson.M{document.IDKey(): bson.M{"$eq": document.ID()}}

	log.Println("filter", filter)

	var data bson.M
	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, result.Err()
	}

	err := result.Decode(&data)

	if err != nil {
		return nil, err
	}

	return D(document.Collection(), data), nil

}

func New(dbName string) Repo {
	return &mrepo{isInitialized: false, dbName: dbName}
}

func getDBURL() string {
	dbURL := os.Getenv("MONGODB_URL")
	if dbURL != "" {
		return dbURL
	}
	return "mongodb://localhost:27017"
}

func (mrepo *mrepo) IsInitialized() bool {
	return mrepo.isInitialized
}

func (mrepo *mrepo) Init() error {
	if mrepo.IsInitialized() {
		return errors.New("repository already initialized")
	}

	url := getDBURL()
	log.Println("connect to", url)

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(getDBURL()))

	if err != nil {
		return err
	}

	mrepo.client = client

	mrepo.isInitialized = true

	return nil
}

func (mrepo *mrepo) Done() {
	err := mrepo.client.Disconnect(context.TODO())

	if err != nil {
		log.Fatal(err)
	}
}
