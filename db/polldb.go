// Package db contain operations for database
package db

import (
	"crypto/sha1"
	"encoding/hex"
	"time"

	"github.com/miska12345/DDPoll/poll"
	goLogger "github.com/phachon/go-logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// PollDB represent an instance of Poll's database
type PollDB struct {
	database          *mongo.Database
	publicCollection  *mongo.Collection
	privateCollection *mongo.Collection
	logger            *goLogger.Logger
	db                *DB
}

func (pb *PollDB) GeneratePID(args ...string) string {
	h := sha1.New()
	for _, v := range args {
		h.Write([]byte(v))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// CreatePoll create a new poll and return the poll id
func (pb *PollDB) CreatePoll(owner, title, content, catergory string, public bool, duration time.Duration, choices []string) (string, error) {
	ctx, cancel := pb.db.QueryContext()
	defer cancel()
	var collection *mongo.Collection

	// Depending on public or not, put into different collections
	// if public {
	// 	collection = pd.database.Collection(pd.publicCollection)
	// } else {
	// 	collection = pd.database.Collection(pd.privateCollection)
	// }

	collection = pb.publicCollection

	pid := pb.GeneratePID(owner, time.Now().String())
	_, err := collection.InsertOne(ctx, bson.M{
		"_id":        pid, // Change in the future
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
func (pb *PollDB) GetPollByPID(id string) (p *poll.Poll, err error) {
	ctx, cancel := pb.db.QueryContext()
	defer cancel()

	p = new(poll.Poll)
	collection := pb.publicCollection
	filter := bson.M{"_id": id}
	err = collection.FindOne(ctx, filter).Decode(p)
	if err != nil {
		pb.logger.Debug(err.Error())
		return
	}
	pb.logger.Debugf("Found poll id %s", id)
	return
}

func (pb *PollDB) GetPollsByUser(username string) (res []*poll.Poll, err error) {
	ctx, cancel := pb.db.QueryContext()
	defer cancel()

	cur, err := pb.publicCollection.Find(ctx, bson.M{
		"owner": username,
	})
	if err != nil {
		return
	}
	var restmp []*poll.Poll
	for cur.Next(ctx) {
		var poll *poll.Poll
		err = cur.Decode(poll)
		if err != nil {
			pb.logger.Error(err.Error())
			return restmp, err
		}
		res = append(res, poll)
	}
	return restmp, nil
}
