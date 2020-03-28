// Package db contain operations for database
package db

import (
	"fmt"
	"time"

	"github.com/miska12345/DDPoll/poll"
	goLogger "github.com/phachon/go-logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// PollDB represent an instance of Poll's database
type PollDB struct {
	database          *mongo.Database
	publicCollection  string
	privateCollection string
	logger            *goLogger.Logger
	db                *DB
}

// CreatePoll create a new poll and return the poll id
func (pd *PollDB) CreatePoll(owner, title, content, catergory string, public bool, duration time.Duration, choices []string) (string, error) {
	ctx, cancel := pd.db.QueryContext()
	defer cancel()
	var collection *mongo.Collection

	// Depending on public or not, put into different collections
	// if public {
	// 	collection = pd.database.Collection(pd.publicCollection)
	// } else {
	// 	collection = pd.database.Collection(pd.privateCollection)
	// }

	collection = pd.database.Collection(pd.publicCollection)
	res, err := collection.InsertOne(ctx, bson.M{
		"owner":      owner,
		"public":     public,
		"title":      title,
		"content":    content,
		"category":   catergory,
		"choices":    choices,
		"votes":      []uint64{},
		"voteLimit":  1,
		"numVoted":   0,
		"numViewed":  0,
		"numStarred": 0,
		"createTime": time.Now(),
		"endTime":    time.Now().Add(duration),
	})

	if err != nil {
		return "", err
	}
	if str, ok := res.InsertedID.(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("Cannot convert resID to string")
}

// GetPollByID return a poll struct
// Currently only support search public poll by id
func (pd *PollDB) GetPollByID(id string) (p *poll.Poll, err error) {
	ctx, cancel := pd.db.QueryContext()
	defer cancel()

	p = new(poll.Poll)
	collection := pd.database.Collection(pd.publicCollection)
	filter := bson.M{"_id": id}
	err = collection.FindOne(ctx, filter).Decode(p)
	if err != nil {
		pd.logger.Debug(err.Error())
		return
	}
	pd.logger.Debugf("Found poll id %s", id)
	return
}
