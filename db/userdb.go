package db

import (
	"crypto/sha1"
	"encoding/hex"
	"time"

	"github.com/miska12345/DDPoll/polluser"
	goLogger "github.com/phachon/go-logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserDB represent an instance of users' database
type UserDB struct {
	database          *mongo.Database
	publicCollection  *mongo.Collection
	privateCollection *mongo.Collection
	logger            *goLogger.Logger
	db                *DB
}

//GenerateUID is a method that generates a unique user id for the userDB
func (ub *UserDB) GenerateUID(args ...string) string {
	h := sha1.New()
	for _, v := range args {
		h.Write([]byte(v))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (ub *UserDB) genRandomBytes(size int) (salt []byte) {
	salt = make([]byte, size)
	//TODO: generate random bytes and put into the array
	return salt
}

//CreateNewUser is a method that creates a new user in the user database
//@Retrun the unique user id
//@Return error if there is any
func (ub *UserDB) CreateNewUser(username, password string) (string, error) {

	ctx, cancel := ub.db.QueryContext()
	defer cancel()

	var collection *mongo.Collection = ub.publicCollection
	var uid string = ub.GenerateUID(username, time.Now().String())
	var passbytes []byte = []byte(password)
	var salt []byte = ub.genRandomBytes(64)

	_, err := collection.InsertOne(ctx, bson.M{
		"_id":  uid,
		"name": username,
		"pass": passbytes,
		"salt": salt,
	})

	if err != nil {
		return "", err
	}

	return uid, nil
}

//GetUserByID will return the user with the id specifield
func (ub *UserDB) GetUserByID(uid string) (u *polluser.User, err error) {
	ctx, cancel := ub.db.QueryContext()
	defer cancel()

	u = new(polluser.User)
	collection := ub.publicCollection
	filter := bson.M{"_id": uid}
	err = collection.FindOne(ctx, filter).Decode(u)

	if err != nil {
		ub.logger.Debug(err.Error())
		return
	}

	ub.logger.Debugf("Found user id %s", u.UID)
	return

}

//GetUserByName will return the user with the name specifield
func (ub *UserDB) GetUserByName(name string) (u *polluser.User, err error) {
	ctx, cancel := ub.db.QueryContext()
	defer cancel()

	u = new(polluser.User)
	collection := ub.publicCollection

	err = collection.FindOne(ctx, bson.M{"name": name}).Decode(u)

	if err != nil {
		ub.logger.Debug(err.Error())
		return
	}

	ub.logger.Debugf("Found user name %s", u.Name)
	return
}
