package db

import (
	"context"

	"github.com/procrastination-team/lamp.api/pkg/config"
	"github.com/procrastination-team/lamp.api/pkg/lamp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func New(conf *config.DatabaseConfig, ctx context.Context) (*Storage, error) {
	clientOptions := options.Client().ApplyURI(conf.Address)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	mongoSt := &Storage{
		client:     client,
		collection: client.Database("procrastination").Collection("lamps"),
	}

	return mongoSt, nil
}

func (s *Storage) GetLamps() ([]lamp.Lamp, error) {
	var lamps []lamp.Lamp
	cur, err := s.collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}

	err = cur.All(context.Background(), &lamps)
	return lamps, err
}
