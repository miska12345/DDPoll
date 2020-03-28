package db

import (
	"testing"
	"time"

	"github.com/miska12345/DDPoll/db"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

const TEST_DB = "test"
const TEST_COLLECTION = "testCollection"

func TestBasicDB(t *testing.T) {
	db, err := Dial("mongodb+srv://admin:wassup@cluster0-n0w7a.mongodb.net/test?retryWrites=true&w=majority", 2*time.Second, 5*time.Second)
	assert.Nil(t, err)
	defer db.Disconnect()

	err = wipeDatabase(db)
	assert.Nil(t, err)

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
	_, err := initializeTestEnv()
	assert.Nil(t, err)

}

func initializeTestEnv() (db *db.DB, err error) {
	db, err = Dial("mongodb+srv://admin:wassup@cluster0-n0w7a.mongodb.net/test?retryWrites=true&w=majority", 2*time.Second, 5*time.Second)
	defer db.Disconnect()

	err = wipeDatabase(db)
	return
}

func wipeDatabase(db *DB) error {
	ctx, cancel := db.QueryContextEx(5 * time.Second)
	defer cancel()

	_, err := db.Client.Database(TEST_DB).Collection(TEST_COLLECTION).DeleteMany(ctx, bson.M{})
	return err
}
