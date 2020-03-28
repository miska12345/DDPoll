package db

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

const DB_LINK = "mongodb+srv://ddpoll:ddpoll@test-ycw1l.mongodb.net/test?retryWrites=true&w=majority"
const TEST_DB = "test"
const TEST_COLLECTION = "testCollection"

func TestUserDB(t *testing.T) {
	db, err := initializeTestEnv()
	defer db.Disconnect()

	ctx, cancel := db.QueryContext()
	defer cancel
	collection := db.Client.Database(TEST_DB).Collection(TEST_COLLECTION)
}

func TestBasicDB(t *testing.T) {
	db, err := initializeTestEnv()
	defer db.Disconnect()

	ctx, cancel := db.QueryContext()
	defer cancel()
	collection := db.Client.Database(TEST_DB).Collection(TEST_COLLECTION)

	_, err = collection.InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159, "desc": "I ate that pie yesterday!"})
	assert.Nil(t, err)

	ctx, cancel2 := db.QueryContext()
	defer cancel2()
	var result struct {
		Name  string
		Value float64
	}
	singRes := collection.FindOne(ctx, bson.M{"name": "pi", "value": 3.14159})
	singRes.Decode(&result)
	assert.Equal(t, result.Name, "pi")
	assert.Equal(t, result.Value, 3.14159)
}

func TestPollsDB(t *testing.T) {
	db, err := initializeTestEnv()
	defer db.Disconnect()
	assert.Nil(t, err)

	pollsDB := db.ToPollsDB(TEST_DB, TEST_COLLECTION, "")
	id, err := pollsDB.CreatePoll("miska", "example poll", "vote for dinner", "Life Style", true, time.Hour, []string{"Chicken", "Rice"})

	fmt.Println(id)
	assert.Nil(t, err)

	// Find the poll with the id
	p, err := pollsDB.GetPollByPID(id)
	assert.Nil(t, err)

	assert.Equal(t, p.PID, id)
	assert.Equal(t, p.Owner, "miska")
	assert.Equal(t, p.Choices, []string{"Chicken", "Rice"})
	assert.Equal(t, p.Votes, []uint64{0, 0})

	// Find the poll with invalid id
	p, err = pollsDB.GetPollByPID("")
	assert.NotNil(t, err)

	res, err := pollsDB.GetPollsByUser("miska")
	assert.Equal(t, 1, len(res))
	assert.Equal(t, res[0].PID, p.PID)
}

func initializeTestEnv() (db *DB, err error) {
	db, err = Dial(DB_LINK, 2*time.Second, 5*time.Second)

	err = wipeDatabase(db)
	return
}

func wipeDatabase(db *DB) error {
	ctx, cancel := db.QueryContextEx(5 * time.Second)
	defer cancel()

	_, err := db.Client.Database(TEST_DB).Collection(TEST_COLLECTION).DeleteMany(ctx, bson.M{})
	return err
}
