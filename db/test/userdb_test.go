package dbtest

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const collectionname = "users"

func TestConcurrentCreateUser(t *testing.T) {
	db, err := initializeTestEnv(collectionname)
	defer db.Disconnect()

	assert.Nil(t, err)

	userDB := db.ToUserDB(Database, collectionname, "")

	a := func(username, password string, wg *sync.WaitGroup) error {
		defer wg.Done()
		if _, err := userDB.CreateNewUser(username, password); err != nil {
			return err
		}
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(6)
	for i := 0; i < 4; i++ {
		a("didntpay4", string(i), &wg)

	}
	a("didntpayyy", string(14), &wg)
	a("didntpay4", string(14), &wg)
	wg.Wait()
}

func TestUpdateUserPoll(t *testing.T) {
	db, err := initializeTestEnv(collectionname)
	defer db.Disconnect()

	assert.Nil(t, err)
	userDB := db.ToUserDB(Database, collectionname, "")
	_, err = userDB.CreateNewUser("didntpay", "password")
	assert.Nil(t, err)
	assert.Nil(t, userDB.UpdateUserPolls("didntpay", "a", 1))
	assert.Nil(t, userDB.UpdateUserPolls("didntpay", "b", 1))
	res, err := userDB.GetUserPollsByGroup("didntpay", 1)
	assert.Nil(t, err)
	assert.Equal(t, []string{"a"}, res)
}
