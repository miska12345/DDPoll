package dbtest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const collectionname = "testCollection_weifeng"

func testConcurrentCreateUser(t *testing.T) {
	db, err := initializeTestEnv(collectionName)
	defer db.Disconnect()

	assert.Nil(t, err)

	userDB := db.ToUsersDB(Database, collectionname, "")

}
