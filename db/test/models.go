package dbtest

import (
	"time"

	"github.com/miska12345/DDPoll/db"
	"go.mongodb.org/mongo-driver/bson"
)

const Database = "testDB"
const DBlink = "mongodb+srv://ddpoll:ddpoll@test-ycw1l.mongodb.net/test?retryWrites=true&w=majority"

func initializeTestEnv(collectionName string) (dbr *db.DB, err error) {
	dbr, err = db.Dial(DBlink, 2*time.Second, 5*time.Second)
	if err != nil {
		return
	}
	err = wipeDatabase(dbr, collectionName)
	return
}

func wipeDatabase(db *db.DB, collectionName string) error {
	ctx, cancel := db.QueryContextEx(5 * time.Second)
	defer cancel()

	_, err := db.Client.Database(Database).Collection(collectionName).DeleteMany(ctx, bson.M{})
	return err
}
