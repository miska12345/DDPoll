package dbtest

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const collectionName = "poll"

func TestPollsDB(t *testing.T) {
	db, err := initializeTestEnv(collectionName)
	defer db.Disconnect()
	assert.Nil(t, err)

	pollsDB := db.ToPollsDB(Database, collectionName, "")
	id, err := pollsDB.CreatePoll("miska", "example poll", "vote for dinner", "Life Style", true, time.Hour, []string{"Chicken", "Rice"})

	fmt.Println(id)
	assert.Nil(t, err)

	// Find the poll with the id
	p, err := pollsDB.GetPollByPID(id)
	assert.Nil(t, err)

	assert.Equal(t, p.PID, id)
	assert.Equal(t, p.Owner, "miska")
	assert.Equal(t, p.Choices, []string{"Chicken", "Rice"})
	assert.Equal(t, p.Votes, []uint64{0, 0})
	assert.Equal(t, 2, len(p.Votes))

	// Find the poll with invalid id
	p, err = pollsDB.GetPollByPID("")
	assert.NotNil(t, err)

	res, err := pollsDB.GetPollsByUser("miska")
	assert.Nil(t, err)
	assert.Equal(t, id, (<-res).PID)

	_, ok := <-res
	assert.False(t, ok)
}

func TestPollsDBNewstPolls(t *testing.T) {
	db, err := initializeTestEnv(collectionName)
	defer db.Disconnect()
	assert.Nil(t, err)
	pollsDB := db.ToPollsDB(Database, collectionName, "")

	ids := make([]string, 10)
	for i := 0; i < 10; i++ {
		id, err := pollsDB.CreatePoll("miska", strconv.Itoa(i), "vote for dinner", "Life Style", true, time.Hour, []string{"Chicken", "Rice"})
		assert.Nil(t, err)
		ids[i] = id
	}
	ch, err := pollsDB.GetNewestPolls(10)
	assert.Nil(t, err)

	for i := 9; i >= 0; i-- {
		val, ok := <-ch
		assert.True(t, ok)
		assert.Equal(t, val.PID, ids[i])
		assert.Equal(t, strconv.Itoa(i), val.Title)
		assert.Equal(t, "miska", val.Owner)
	}
	_, ok := <-ch
	assert.False(t, ok)
}
