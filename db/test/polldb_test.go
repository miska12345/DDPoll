package dbtest

import (
	"fmt"
	"strconv"
	"sync"
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
	ch, err := pollsDB.GetNewestPolls(5)
	assert.Nil(t, err)

	for i := 9; i >= 5; i-- {
		val, ok := <-ch
		assert.True(t, ok)
		assert.Equal(t, val.PID, ids[i])
		assert.Equal(t, strconv.Itoa(i), val.Title)
		assert.Equal(t, "miska", val.Owner)
	}
	_, ok := <-ch
	assert.False(t, ok)
}

func TestFindPollsUser(t *testing.T) {
	db, err := initializeTestEnv(collectionName)
	defer db.Disconnect()
	assert.Nil(t, err)

	ids := make([]string, 10)
	pollsDB := db.ToPollsDB(Database, collectionName, "")
	for i := 0; i < 10; i++ {
		id, err := pollsDB.CreatePoll("miska", "title is not important", "", "", true, time.Hour, []string{"Yes", "No"})
		assert.Nil(t, err)
		ids[i] = id
		_, err = pollsDB.CreatePoll("not miska", "title is very important", "", "", true, time.Hour, []string{"Yes", "No"})
		assert.Nil(t, err)
	}

	ch, err := pollsDB.GetPollsByUser("miska")
	assert.Nil(t, err)

	for i := 0; i < 10; i++ {
		v, ok := <-ch
		assert.True(t, ok)
		assert.Equal(t, "miska", v.Owner)
		assert.Contains(t, ids, v.PID)
	}
	_, ok := <-ch
	assert.False(t, ok)
}

func TestBasicPollUpdate(t *testing.T) {
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
	sampleVotes := []uint64{1, 2}
	for i := 0; i < 10; i++ {
		pollsDB.UpdateNumVoted(ids[i], sampleVotes)
	}
	for i := 0; i < 10; i++ {
		p, err2 := pollsDB.GetPollByPID(ids[i])
		assert.Nil(t, err2)
		assert.Equal(t, sampleVotes, p.Votes)
	}
}

func TestConcurrentUpdate(t *testing.T) {
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
	var wg sync.WaitGroup
	threads := 200
	wg.Add(threads)
	a := func(ids []string, wg *sync.WaitGroup) {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			err := pollsDB.UpdateNumVoted(ids[i], []uint64{1, 1})
			assert.Nil(t, err)
		}
	}
	for i := 0; i < threads; i++ {
		go a(ids, &wg)
	}
	wg.Wait()
	expect := []uint64{uint64(threads), uint64(threads)}
	for i := 0; i < 10; i++ {
		p, err2 := pollsDB.GetPollByPID(ids[i])
		assert.Nil(t, err2)
		assert.Equal(t, expect, p.Votes)
		assert.Equal(t, uint64(threads), p.NumVoted)
	}
}

func TestAddStar(t *testing.T) {
	db, err := initializeTestEnv(collectionName)
	defer db.Disconnect()
	assert.Nil(t, err)
	pollsDB := db.ToPollsDB(Database, collectionName, "")

	id, err := pollsDB.CreatePoll("miska", "title", "content", "cat", true, time.Hour, []string{"A", "B"})
	assert.Nil(t, err)

	// Add 10 stars sequentially
	for i := 0; i < 10; i++ {
		err = pollsDB.AddPollStar(id)
		assert.Nil(t, err)
	}

	p, err := pollsDB.GetPollByPID(id)
	assert.Nil(t, err)
	assert.Equal(t, uint64(10), p.Stars)

	id2, err := pollsDB.CreatePoll("miska", "title", "content", "cat", true, time.Hour, []string{"A", "B"})
	assert.Nil(t, err)
	// Add 10 stars in parallel
	threads := 1000
	var wg sync.WaitGroup
	wg.Add(threads)
	for i := 0; i < threads; i++ {
		go func(id string, w *sync.WaitGroup) {
			err := pollsDB.AddPollStar(id)
			assert.Nil(t, err)
			w.Done()
		}(id2, &wg)
	}
	wg.Wait()
	p, err = pollsDB.GetPollByPID(id2)
	assert.Nil(t, err)
	assert.Equal(t, uint64(threads), p.Stars)
}
