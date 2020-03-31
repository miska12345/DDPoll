package dbtest

import (
	"math/rand"
	"sync"
	"testing"
	"time"

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

	asserted := make(map[uint32][]string)
	for i := 0; i < 500; i++ {
		// x := rand.Intn(3) + 1
		// switch x {
		// case 1:
		pid := userDB.GenerateUID("didntpay", time.Now().String())
		groupID := uint32(rand.Int())
		assert.Nil(t, userDB.UpdateUserPolls("didntpay", pid, groupID))
		asserted[groupID] = append(asserted[groupID], pid)
		// case 2:

		// }
	}

	for key, val := range asserted {
		ids, err := userDB.GetUserPollsByGroup("didntpay", key)
		assert.Nil(t, err)

		for i := range ids {
			assert.Equal(t, ids[i], val[i])
		}

	}
}

// func TestConcurrentUserPoll (t *testing.T) {

// }
