package db

import (
	"testing"

	"github.com/miska12345/DDPoll/db"
	"github.com/stretchr/testify/assert"
)

func TestBasicDB(t *testing.T) {
	db, err := db.Dial("mongodb+srv://admin:wassup@cluster0-n0w7a.mongodb.net/test?retryWrites=true&w=majority")
	assert.Nil(t, err)

	pollsDB := db.ToPollsDB("Test")
	_ = pollsDB
}
