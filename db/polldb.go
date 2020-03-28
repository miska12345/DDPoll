// Package db contain operations for database
package db

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// PollDB represent an instance of Poll's database
type PollDB struct {
	database *mongo.Database
	clName   string
	db       *DB
}

// CreatePoll create a new poll
func (pd *PollDB) CreatePoll(host, title, content string, public bool, choices []string) (string, error) {
	panic("unimplemented")
}
