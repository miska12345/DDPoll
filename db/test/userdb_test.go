package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const collectionname = "testCollection_weifeng"

func testConcurrentCreateUser(t *testing.T) {
	db, err := initializeTestEnv(collectionName)
	defer db.Disconnect()

	assert.Nil(err)

	userDB := db.ToUsersDB(Database, collectionname, "")

}
