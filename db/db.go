// Package db provides a generic interface for database operations
package db

import (
	"context"
	"time"

	goLogger "github.com/phachon/go-logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var logger *goLogger.Logger

// DB represent an instance of database
type DB struct {
	Client       *mongo.Client
	queryTimeout time.Duration
}

// Dial connect to a database server and return Database instance
func Dial(URL string, connectionTimeout time.Duration, queryTimeout time.Duration) (db *DB, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URL))
	if err != nil {
		return
	}

	ctx, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel2()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return
	}

	// We are connected
	ctx, cancel3 := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel3()
	db = &DB{
		Client:       client,
		queryTimeout: queryTimeout,
	}

	logger = goLogger.NewLogger()
	logger.Detach("console")

	// console adapter config
	consoleConfig := &goLogger.ConsoleConfig{
		Color:      true,  // Does the text display the color
		JsonFormat: false, // Whether or not formatted into a JSON string
		Format:     "",    // JsonFormat is false, logger message output to console format string
	}

	// add output to the console
	logger.Attach("console", goLogger.LOGGER_LEVEL_DEBUG, consoleConfig)
	return
}

// SetQueryTimeOut will update the queryTimeout to timeout
func (d *DB) SetQueryTimeOut(timeout time.Duration) {
	d.queryTimeout = timeout
}

// ToPollsDB convert the current DB instance to a PollDB instance
func (d *DB) ToPollsDB(database, publicCollectionName, privateCollectionName string) *PollDB {
	db := d.Client.Database(database)
	return &PollDB{
		publicCollection: db.Collection(publicCollectionName),
		logger:           logger,
		db:               d,
	}
}

// ToUsersDB convert the current DB instance to a UsersDB instance
func (d *DB) ToUserDB(database, publicCollectionName, privateCollectionName string) *PollDB {
	db := d.Client.Database(database)
	return &UsersDB{
		database:         db,
		publicCollection: db.Collection(publicCollectionName),
		logger:           logger,
		db:               d,
	}
}

// Disconnect will disconnect the current client
func (d *DB) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	d.Client.Disconnect(ctx)
}

// QueryContext will return a timeout context with queryTimeout as the timeout
func (d *DB) QueryContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d.queryTimeout)
}

// QueryContextEx will return a timeout context with timeout as the timeout
func (d *DB) QueryContextEx(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
