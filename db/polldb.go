// Package db contain operations for database
package db

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"time"

	"github.com/miska12345/DDPoll/poll"
	goLogger "github.com/phachon/go-logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PollDB represent an instance of Poll's database
type PollDB struct {
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
		"createTime": time.Now().UTC(),
		"endTime":    time.Now().Add(duration).UTC(),
	})

	if err != nil {
		return "", err
	}
	return pid, nil
}

// GetPollByPID return a poll struct by the provided pid
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

// GetPollsByUser return all polls created by the user
func (pb *PollDB) GetPollsByUser(username string) (ch chan *poll.Poll, err error) {
	ctx, cancel := pb.db.QueryContext()
	defer cancel()

	cur, err := pb.publicCollection.Find(ctx, bson.M{
		"owner": username,
	})
	if err != nil {
		return
	}
	ch = make(chan *poll.Poll)
	go func(ch chan *poll.Poll, c *mongo.Cursor) {
		for cur.Next(ctx) {
			var poll poll.Poll
			err = cur.Decode(&poll)
			if err != nil {
				pb.logger.Error(err.Error())
				close(ch)
				break
			}
			ch <- &poll
		}
		close(ch)
	}(ch, cur)
	return
}

// GetNewestPolls return at most 'count' number of polls, sorted by create time
func (pb *PollDB) GetNewestPolls(count int64) (ch chan *poll.Poll, err error) {
	ctx, cancel := pb.db.QueryContext()
	defer cancel()

	findOption := options.Find()
	findOption.SetSort(bson.M{
		"createTime": -1,
	})
	findOption.SetLimit(count)
	cur, err := pb.publicCollection.Find(ctx, bson.M{}, findOption)
	if err != nil {
		return
	}
	ch = make(chan *poll.Poll)
	go func(ch chan *poll.Poll, c *mongo.Cursor) {
		for cur.Next(ctx) {
			var poll poll.Poll
			err = cur.Decode(&poll)
			if err != nil {
				pb.logger.Error(err.Error())
				close(ch)
				break
			}
			ch <- &poll
		}
		close(ch)
	}(ch, cur)
	return
}

// AddPollStar add a star to the given poll with id pollID
func (pb *PollDB) AddPollStar(pollID string) (err error) {
	ctx := context.Background()

	//defer cancel()
	_, err = pb.publicCollection.UpdateOne(ctx, bson.M{
		"_id": pollID,
	}, bson.M{
		"$inc": bson.M{
			"stars": 1,
		},
	})
	return
}

// TODO: Change function name to update votes
func (pb *PollDB) UpdateNumVoted(pid string, votes []uint64) (err error) {
	ctx, cancel := pb.db.QueryContext()
	defer cancel()
	session, err := pb.db.Client.StartSession()
	if err != nil {
		return err
	}
	if err = session.StartTransaction(); err != nil {
		return err
	}
	if err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		p, err := pb.GetPollByPID(pid)
		if err != nil {
			sc.AbortTransaction(sc)
			return err
		}
		for idx, _ := range p.Votes {
			p.Votes[idx] += votes[idx]
		}
		_, err2 := pb.publicCollection.UpdateOne(
			sc,
			bson.M{"_id": pid},
			bson.M{"$set": bson.M{"votes": p.Votes}})
		sc.CommitTransaction(sc)
		if err2 != nil {
			pb.logger.Debugf("Error %s on updating poll id %s", err2, pid)
		} else {
			pb.logger.Debugf("Updated poll id %s", pid)
		}
		return err2
	}); err != nil {
		return err
	}

	session.EndSession(ctx)
	return nil
}
