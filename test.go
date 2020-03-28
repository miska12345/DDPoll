package main

import (
	"context"
	"fmt"
	"log"
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
	collection := client.Database("testing").Collection("numbers")
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	res, err := collection.InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
	id := res.InsertedID
	fmt.Println(id)

	var result struct {
		Value float64
	}
	filter := bson.M{"name": "pix", "value": 3.14159}
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Value)

	/*
		deleteDoc := bson.M{"name": "pi"}
		ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
		res2, err := collection.DeleteOne(ctx, deleteDoc)

		fmt.Println(res2.DeletedCount)
	*/

	_, err = collection.UpdateOne(ctx, bson.M{"value": bson.M{"$eq": 666}}, bson.M{"$set": bson.M{"name": "pi", "value": 777}})
	//fmt.Println(udres.ModifiedCount)

	cur, err := collection.Find(ctx, bson.M{})
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var result struct {
			Name  string
			Value float64
		}
		err := cur.Decode(&result)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(result)
	}
}
