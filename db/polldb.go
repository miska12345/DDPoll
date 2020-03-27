// Package db contain operations for database
package db

import "go.mongodb.org/mongo-driver/mongo"

type PollDB struct {
	database *mongo.Database
	clName   string
	db       DB
}

func (pd *PollDB) CreatePoll() {

}
