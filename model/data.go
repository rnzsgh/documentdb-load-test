package model

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	docdb "github.com/rnzsgh/fargate-documentdb-compute-poc/db"
)

const dataTestCount = 1000

const dataTestItems = 2500

type Data struct {
	Id      *primitive.ObjectID `json:"id" bson:"_id"`
	P       []float64           `json:"p" bson:"p"`
	Entropy float64             `json:"entropy" bson:"entropy"`
}

func DataEnsureTest() error {

	if err := DataDeleteAll(); err != nil {
		return fmt.Errorf("Unable to delete all data - reason: %v", err)
	}

	if count, err := DataCount(); err != nil {
		return err
	} else if count == 0 {
		for i := 0; i < dataTestCount; i++ {
			id := primitive.NewObjectID()
			data := &Data{Id: &id}

			data.P = make([]float64, dataTestItems)

			r := rand.New(rand.NewSource(time.Now().UnixNano()))

			for i := 0; i < dataTestItems; i++ {
				data.P[i] = r.Float64()
			}

			if err := DataCreate(data); err != nil {
				return fmt.Errorf("Unable to create test data - reason %v", err)
			}
		}

	} else if count != dataTestCount {
		return fmt.Errorf("Invalid data count - expected: %d - received: %d", dataTestCount, count)
	}
	return nil
}

func DataCount() (count int64, err error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	count, err = dataCollection().Count(ctx, bson.D{})
	return
}

func DataFindAllIds() ([]*primitive.ObjectID, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := dataCollection().Find(ctx, bson.D{}, &options.FindOptions{Projection: bson.D{{"_id", 1}}})

	var ids []*primitive.ObjectID

	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("Failed to decode data: %v", err)
		} else {
			id := doc["_id"].(primitive.ObjectID)
			ids = append(ids, &id)
		}
	}

	return ids, nil
}

func DataDeleteAll() (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	_, err = dataCollection().DeleteMany(ctx, bson.D{})
	return
}

func DataCreate(data *Data) (err error) {
	ctx, _ := context.WithTimeout(context.Background(), 300*time.Second)
	_, err = dataCollection().InsertOne(ctx, data)
	return
}

func DataExists(id *primitive.ObjectID) (bool, error) {
	data, err := DataFindById(id)
	if err != nil {
		return false, err
	}

	if data != nil {
		return true, nil
	}

	return false, nil
}

func DataFindById(id *primitive.ObjectID) (*Data, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	data := &Data{}

	err := dataCollection().FindOne(ctx, bson.D{{"_id", id}}).Decode(data)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return data, err
}

func dataCollection() *mongo.Collection {
	return docdb.Client.Database("work").Collection("data")
}
