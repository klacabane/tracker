package main

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	testfetcher = NewFetcher(3, WEEK, []int{}, []string{"testf"})
	tdata       = timeData{date: time.Now(), period: WEEK}
)

func TestSetKeys(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	testfetcher.setKeys(&wg)
	wg.Wait()

	assert.Equal(t, tdata.Key(), testfetcher.periodKeys[3])
	assert.Equal(t, tdata.Prev().Key(), testfetcher.periodKeys[2])
}

func TestFetchInvalid(t *testing.T) {
	fetcher := NewFetcher(5, DAY, []int{}, []string{"foo", "bar"})
	go fetcher.fetch()

	assert.Equal(t, "all categories", <-fetcher.catnamec)
	for i := 0; i < 2; i++ {
		res := <-fetcher.resc
		_, ok := res.err.(*ErrInvalidDB)
		assert.True(t, ok)
	}
	<-fetcher.quit
}

func TestFetchEmpty(t *testing.T) {
	go testfetcher.fetch()

	assert.Equal(t, "all categories", <-testfetcher.catnamec)

	res := <-testfetcher.resc
	assert.Nil(t, res.err)
	assert.Equal(t, 0, len(res.values))

	<-testfetcher.quit

}

func TestFetchWithData(t *testing.T) {
	err := populatedb()
	assert.Nil(t, err)

	testfetcher.categories = []int{1}
	go testfetcher.fetch()

	assert.Equal(t, "default", <-testfetcher.catnamec)

	res := <-testfetcher.resc
	assert.Nil(t, res.err)
	assert.Equal(t, 1, len(res.values))
	assert.Equal(t, tdata.Key(), res.values[0].Key())
	assert.Equal(t, 17020, res.values[0].Quantity())

	<-testfetcher.quit

}

func TestExec(t *testing.T) {
	err := testfetcher.Exec()
	assert.Nil(t, err)
	assert.Equal(t, 17020, testfetcher.data[tdata.Key()])
	assert.Equal(t, 170.2, testfetcher.Sum())
	assert.Equal(t, "default", testfetcher.CatNames()[0])
}

func populatedb() error {
	err := withDBContext("testf", func(db *DB) error {
		var cerr error
		add := func(qty int64) {
			if cerr != nil {
				return
			}
			cerr = db.addRecord(qty, 1)
		}

		add(1000)
		add(15500)
		add(520)
		return cerr
	})
	return err
}
