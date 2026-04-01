package mongostorage

import (
	"fmt"

	"github.com/fishkaoff/ts-backend/internal/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoStore struct {
	cfg config.MongoConfig

	usersCollection    *mongo.Collection
	productsCollection *mongo.Collection
	cartsCollection    *mongo.Collection
}

func New(cfg config.MongoConfig) *MongoStore {
	return &MongoStore{
		cfg: cfg,
	}
}

func (s *MongoStore) Connect() (*mongo.Client, error) {
	const op = "mongostorage.Connect"
	uri := s.cfg.ClusterUrl

	client, err := mongo.Connect(options.Client().
		ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	s.usersCollection = client.Database("ts").Collection("users")
	s.productsCollection = client.Database("ts").Collection("products")
	s.cartsCollection = client.Database("ts").Collection("carts")
	return client, nil
}
