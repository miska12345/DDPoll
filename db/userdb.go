package db

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/miska12345/DDPoll/polluser"
	goLogger "github.com/phachon/go-logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserDB represent an instance of users' database
type UserDB struct {
	database          *mongo.Database
	publicCollection  *mongo.Collection
	privateCollection *mongo.Collection
	logger            *goLogger.Logger
	db                *DB
}

//var ErrUserNameTaken = errors.New("User name already taken")

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
	rand.Read(salt)
	return salt
}

//CreateNewUser is a method that creates a new user in the user database
//@Retrun the unique user id
//@Return error if there is any
func (ub *UserDB) CreateNewUser(username, password string) (string, error) {
	h := sha1.New()

	var salt []byte = ub.genRandomBytes(64)
	var collection *mongo.Collection = ub.publicCollection
	var uid string = ub.GenerateUID(username, time.Now().String())
	var passhashed []byte

	h.Write([]byte(password))
	h.Write(salt)
	passhashed = h.Sum(nil)

	ctx, cancel := ub.db.QueryContext()
	defer cancel()

	//look for any document that applys
	filter := bson.M{"name": username}

	//if there is none, insert one in the following format
	replace := bson.M{
		"$setOnInsert": bson.M{
			"_id":       uid,
			"name":      username,
			"pass":      passhashed,
			"salt":      salt,
			"pollGroup": bson.D{},
		},
	}

	err := collection.FindOneAndUpdate(ctx, filter, replace, options.FindOneAndUpdate().SetUpsert(true))

	if err.Err() == mongo.ErrNoDocuments {
		//There is no user with this name
		logger.Debug("Created user " + username)
		return uid, nil
	} else if err.Err() != nil {
		//There is user with this name and something went wrong
		return "", err.Err()
	} else {
		//There is user with this name, tell client to pick another one
		return "", status.Error(codes.AlreadyExists, fmt.Sprintf("User with name %s already exist", username))
	}
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

// UpdateUserPolls will record a new poll in user's history
func (ub *UserDB) UpdateUserPolls(username string, pid string, groupID uint32) (err error) {
	ctx, cancel := ub.db.QueryContext()
	defer cancel()

	strgroupID := strconv.Itoa(int(groupID))
	//logger.Debug(strgroupID)
	filter := bson.M{"name": username}

	update := bson.M{
		"$push": bson.M{
			"pollGroup." + strgroupID: pid,
		},
	}

	_, err = ub.publicCollection.UpdateOne(ctx, filter, update)
	return
}

//GetUserPollsByGroup will return the list of poll ids in the given group
func (ub *UserDB) GetUserPollsByGroup(username string, groupID uint32) (res []string, err error) {
	ctx, cancel := ub.db.QueryContext()
	defer cancel()

	var document = bson.M{}
	var ids []string = make([]string, 0)

	filter := bson.M{
		"name": username,
	}

	projection := bson.M{
		"_id":  0,
		"name": 0,
		"pass": 0,
		"salt": 0,
	}

	var doc = ub.publicCollection.FindOne(ctx, filter, options.FindOne().SetProjection(projection))

	if err := doc.Decode(document); err != nil {
		return nil, err
	}

	retrievedGrpsRaw := document["pollGroup"]

	// this is a map, key as group ID, associating with a list of Polls under that group
	retrievedGrps := reflect.ValueOf(retrievedGrpsRaw)

	for _, key := range retrievedGrps.MapKeys() {
		//find the group we are looking for
		if key.String() == strconv.Itoa(int(groupID)) {
			//retrievedGrps.MapIndex(key).Interface() returns an interface that the key
			//points to. And underneath this interface is an array of string
			retrievedIds := reflect.ValueOf(retrievedGrps.MapIndex(key).Interface())

			for i := 0; i < retrievedIds.Len(); i++ {
				//Knowing it is an array under, we can just index it to get elem at that index
				ids = append(ids, retrievedIds.Index(i).Interface().(string))

			}
		}
	}

	return ids, nil
}

//GetUserByName will return the user with the name specifield
func (ub *UserDB) GetUserByName(name string) (u *polluser.User, err error) {
	ctx, cancel := ub.db.QueryContext()
	defer cancel()

	u = new(polluser.User)
	collection := ub.publicCollection

	err = collection.FindOne(ctx, bson.M{"name": name}).Decode(u)

	if err == mongo.ErrNoDocuments {
		return nil, status.Error(codes.NotFound, "User not found")
	} else if err != nil {
		logger.Debug(err.Error() + " at getting user by name")
		return
	}

	//ub.logger.Debugf("Found user name %s", u.Name)
	return
}

//GetUserAuthSaltAndCred returns the salt for the client
func (ub *UserDB) GetUserAuthSaltAndCred(username string) (salt []byte, token []byte, err error) {
	user, getuserErr := ub.GetUserByName(username)
	if getuserErr != nil {
		return nil, nil, getuserErr
	}

	return user.Salt, user.Pass, nil

}
