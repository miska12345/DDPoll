// Package db contain operations for database
package db

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// PollDB represent an instance of Poll's database
type PollDB struct {
	database *mongo.Database
	clName   string
	db       *DB
}

// CreatePoll create a new poll and return the poll id
func (pd *PollDB) CreatePoll(owner, title, content, catergory string, public bool, choices []string) (string, error) {
	ctx, cancel := pd.db.QueryContext()
	defer cancel()
	collection := pd.database.Collection(pd.clName)
	res, err := collection.InsertOne(ctx, bson.M{
		"isPublic":   public,
		"owner":      owner,
		"title":      title,
		"content":    content,
		"category":   catergory,
		"choices":    choices,
		"voteLimit":  1,
		"createTime": time.Now(),
		"endTime":    time.Now().Add(time.Hour),
	})
	if err != nil {
		return "", err
	}
	if str, ok := res.InsertedID.(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("Cannot convert resID to string")
}
