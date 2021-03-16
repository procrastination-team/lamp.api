package db

import (
	"context"

	"github.com/procrastination-team/lamp.api/pkg/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	client *mongo.Client
	db     *mongo.Database
	coll   *mongo.Collection
	opts   *options.DeleteOptions
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
	db := client.Database(conf.Database)

	mongoSt := &Storage{
		client: client,
		db:     db,
		coll:   client.Database(conf.Database).Collection(conf.Collection),
		opts: options.Delete().SetCollation(&options.Collation{
			Locale:    "en_US",
			Strength:  1,
			CaseLevel: false,
		}),
	}

	return mongoSt, nil
}
