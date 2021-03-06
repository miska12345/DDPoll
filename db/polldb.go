// Package db contain operations for database
package db

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"strconv"
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
	ch, err = pb.GetPolls(count, "createTime", -1)
	if err != nil {
		pb.logger.Error(err.Error())
	}
	return
}

// GetMostViewedPolls return at most "count" number of polls, sorted by views
func (pb *PollDB) GetMostViewedPolls(count int64) (ch chan *poll.Poll, err error) {
	ch, err = pb.GetPolls(count, "numViewed", -1)
	if err != nil {
		pb.logger.Error(err.Error())
	}
	return
}

// GetMostStarPolls return at most "count" number of polls, sorted by number of stars
func (pb *PollDB) GetMostStarPolls(count int64) (ch chan *poll.Poll, err error) {
	ch, err = pb.GetPolls(count, "numStarred", -1)
	if err != nil {
		pb.logger.Error(err.Error())
	}
	return
}

// GetPolls return at most "count" number of polls, sorted by the field "sortBy" in indx order
func (pb *PollDB) GetPolls(count int64, sortBy string, indx int) (ch chan *poll.Poll, err error) {
	ctx := context.Background()
	findOption := options.Find()
	findOption.SetSort(bson.M{
		sortBy: indx,
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
			"numStarred": 1,
		},
	})
	return
}

// TODO: Change function name to update votes
func (pb *PollDB) UpdateNumVoted(pid string, votes []uint64) (err error) {
	ctx := context.Background()
	if err != nil {
		return
	}
	m := make(map[string]uint64)
	// Construct an update map for votes
	for idx, val := range votes {
		if val != 0 {
			target := "votes." + strconv.Itoa(idx)
			m[target] = val
		}
	}
	// Increment number of poll participants
	m["numVoted"] = 1
	_, err = pb.publicCollection.UpdateOne(
		ctx,
		bson.M{
			"_id": pid,
		}, bson.M{
			"$inc": m,
		})
	if err != nil {
		pb.logger.Errorf("%s", err)
	}
	return
}

// AddPollViewCount add one view count to the specific poll
func (pb *PollDB) AddPollViewCount(pid string) (err error) {
	ctx := context.Background()
	_, err = pb.publicCollection.UpdateOne(ctx, bson.M{
		"_id": pid,
	}, bson.M{
		"$inc": bson.M{
			"numViewed": 1,
		},
	})
	return
}

func (pb *PollDB) DeletePollByPID(pid string) (err error) {
	ctx, cancel := pb.db.QueryContext()
	defer cancel()

	_, err = pb.publicCollection.DeleteOne(ctx, bson.M{
		"_id": pid,
	})
	return
}
