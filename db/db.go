// Package db provides a generic interface for database operations
package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Type DB represent an instance of database
type DB struct {
	client       *mongo.Client
	queryTimeout uint
}

// Dial connect to a database server and return Database instance
func Dial(URL string, connectionTimeout uint, queryTimeout uint) (db *DB, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(connectionTimeout)*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URL))
	if err != nil {
		return
	}

	ctx, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return
	}

	// We are connected
	ctx, cancel3 := context.WithTimeout(context.Background(), time.Duration(queryTimeout)*time.Second)
	defer cancel3()
	db = &DB{
		client:       client,
		queryTimeout: queryTimeout,
	}
	return
}

func (d *DB) SetQueryTimeOut(timeout uint) {
	d.queryTimeout = timeout
}

func (d *DB) ToPollsDB(database string, collectionName string) *PollDB {
	return &PollDB{
		database: d.client.Database(database),
		clName:   collectionName,
	}
}
