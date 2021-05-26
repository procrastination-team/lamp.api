package db

import (
	"context"

	"github.com/procrastination-team/lamp.api/pkg/config"
	"github.com/procrastination-team/lamp.api/pkg/format"
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
		collection: client.Database(conf.Database).Collection(conf.Collection),
	}

	return mongoSt, nil
}

func (s *Storage) GetLamps() ([]format.Lamp, error) {
	var lamps []format.Lamp
	cur, err := s.collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}

	err = cur.All(context.Background(), &lamps)
	return lamps, err
}

func (s *Storage) GetLampByID(id string) (format.Lamp, error) {
	var l format.Lamp
	err := s.collection.FindOne(context.TODO(), bson.M{"id": id}, options.FindOne()).Decode(&l)
	return l, err
}

func (s *Storage) CreateLamp(lamp format.Lamp) error {
	_, err := s.collection.InsertOne(context.TODO(), lamp)
	return err
}

func (s *Storage) UpdateLamp(lamp format.Lamp) error {
	_, err := s.collection.UpdateOne(context.TODO(), bson.M{"id": lamp.ID}, bson.M{"$set": lamp}, options.Update().SetUpsert(true))
	return err
}

func (s *Storage) DeleteLamp(id string) error {
	_, err := s.collection.DeleteOne(context.TODO(), bson.M{"id": id}, options.Delete())
	return err
}
