// Package db provides an interface for MongoDB operations
package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Type DB represent an instance of MongoDB
type DB struct {
	client       *mongo.Client
	queryContext context.Context
}

// Dial connect to a database server and return Database instance
func Dial(URL string, connectionTimeout uint, queryTimeout uint) (db *DB, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(connectionTimeout)*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URL))
	if err != nil {
		return
	}

	ctx, _ = context.WithTimeout(context.Background(), 2*time.Second)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return
	}

	// We are connected
	ctx, _ = context.WithTimeout(context.Background(), time.Duration(queryTimeout)*time.Second)
	db = &DB{
		client:       client,
		queryContext: ctx,
	}
}

func (d *DB) SetQueryTimeOut(timeout uint) {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	d.queryContext = ctx
}

func (d *DB) 