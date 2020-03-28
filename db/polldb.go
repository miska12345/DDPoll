// Package db contain operations for database
package db

import (
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

	pid := owner + "#1"
	_, err := collection.InsertOne(ctx, bson.M{
		"pid":        pid, // Change in the future
		"owner":      owner,
		"public":     public,
		"title":      title,
		"content":    content,
		"category":   catergory,
		"choices":    choices,
		"votes":      make([]uint64, len(choices)),
		"voteLimit":  uint64(1),
		"numVoted":   uint64(0),
		"numViewed":  uint64(0),
		"numStarred": uint64(0),
		"createTime": time.Now(),
		"endTime":    time.Now().Add(duration),
	})

	if err != nil {
		return "", err
	}
	return pid, nil
}

// GetPollByID return a poll struct
// Currently only support search public poll by id
func (pd *PollDB) GetPollByPID(id string) (p *poll.Poll, err error) {
	ctx, cancel := pd.db.QueryContext()
	defer cancel()

	p = new(poll.Poll)
	collection := pd.database.Collection(pd.publicCollection)
	filter := bson.M{"pid": id}
	err = collection.FindOne(ctx, filter).Decode(p)
	if err != nil {
		pd.logger.Debug(err.Error())
		return
	}
	pd.logger.Debugf("Found poll id %s", id)
	return
}
