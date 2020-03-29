package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb+srv://admin:wassup@cluster0-n0w7a.mongodb.net/test?retryWrites=true&w=majority",
	))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database("testDB").Collection("transactionTest")
	collection.InsertOne(ctx, bson.M{
		"docid":    1,
		"changeMe": 0,
	})

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(c *mongo.Collection, w *sync.WaitGroup) {
			c.UpdateOne(context.Background(), bson.M{
				"docid": 1,
			}, bson.M{
				"$inc": bson.M{
					"changeMe": 1,
				},
			})
			w.Done()
		}(collection, &wg)
	}
	wg.Wait()
	fmt.Println("Done 1")

	collection.InsertOne(ctx, bson.M{
		"docid":         1,
		"changeMeArray": []int{0, 0, 0},
	})

	var wg2 sync.WaitGroup
	wg2.Add(10)
	m := make(map[string]uint64)
	m["changeMeArray.1"] = 2
	for i := 0; i < 10; i++ {
		go func(c *mongo.Collection, w *sync.WaitGroup) {
			c.UpdateOne(context.Background(), bson.M{
				"docid": 1,
			}, bson.M{
				"$inc": m,
			})
			w.Done()
		}(collection, &wg2)
	}
	wg2.Wait()
	fmt.Println("Done 2")
}
