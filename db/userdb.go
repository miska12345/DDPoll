package db

import (
	"Time"
	"crypto/sha1"
	"encoding/hex"

	goLogger "github.com/phachon/go-logger"
	"github.com/miska12345/user/User"

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
	for _, val := range salt {
		//TODO: generate random bytes and put into the array
	}
	return salt
}

//CreateNewUser is a method that creates a new user in the user database
//@Retrun the unique user id
//@Return error if there is any
func (ub *UserDB) CreateNewUser(username, password string) {

	ctx, cancel := ub.db.QueryContext()
	defer cancel()

	var collection *mongo.Collection = ub.publicCollection
	var uid string = ub.GenerateUID(username, password, Time.Now())
	var passbytes []byte = []byte(password)
	var salt []byte = ub.genRandomBytes(64)

	_, err := collection.InsertOne(ctx, bson.M{
		"_id":  uid,
		"name": username,
		"pass": password,
		"salt": salt,
	})

	if err != nil {
		return "", err
	}

	return uid, nil
}

func (ub *UserDB) GetUserByID(uid int) (u *User) {
	u *
}

func (ub *UserDB) GetUserByName(name string)
