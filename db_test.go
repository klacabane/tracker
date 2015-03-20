package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	testDB    *DB
	dbtest    = dbpath("test")
	dbfetcher = dbpath("testf")

	date       = time.Now()
	year, week = date.ISOWeek()
	month      = int(date.Month())
)

func TestMain(m *testing.M) {
	defer os.Remove(dbtest)
	defer os.Remove(dbfetcher)

	TRACKER_DIR, _ = os.Getwd()

	m.Run()
}

func TestDB(t *testing.T) {
	assert.False(t, exists(dbtest))
	assert.Nil(t, create(dbtest))
	assert.Nil(t, create(dbfetcher))
	assert.True(t, exists(dbtest))

	db, err := open(dbtest)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	assert.Nil(t, db.Close())
}

func TestDblist(t *testing.T) {
	names, err := dblist()
	assert.Nil(t, err)

	assert.Equal(t, 2, len(names))
	assert.Equal(t, "test", names[0])
}

func TestCategories(t *testing.T) {
	testDB, _ = open(dbtest)

	categories, err := testDB.getCategories()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(categories))
	assert.Equal(t, "default", categories[1])

	category, err := testDB.getCategory(1)
	assert.Nil(t, err)
	assert.Equal(t, "default", category)

	category, err = testDB.getCategory(2)
	assert.Equal(t, ErrInvalidCategory, err)
	assert.Equal(t, "", category)

	err = testDB.addCategories("foo", "bar", "baz")
	assert.Nil(t, err)

	categories, err = testDB.getCategories()
	assert.Nil(t, err)
	assert.Equal(t, 4, len(categories))
}

func TestAddRecord(t *testing.T) {
	err := testDB.addRecord(1200, 2)
	assert.Nil(t, err)
}

func TestQueryWeek(t *testing.T) {
	datas, err := testDB.queryWeek(0, []int{2})
	assert.Nil(t, err)

	assert.Equal(t, 1, len(datas))
	assert.Equal(t, 1200, datas[0].Quantity())
	assert.Equal(t, fmt.Sprintf("W%02d %d", week, year), datas[0].Key())

	datas, err = testDB.queryWeek(0, []int{3})
	assert.Nil(t, err)

	assert.Equal(t, 0, len(datas))
}

func TestQueryMonth(t *testing.T) {
	datas, err := testDB.queryMonth(2, []int{2})
	assert.Nil(t, err)

	assert.Equal(t, 1, len(datas))
	assert.Equal(t, 1200, datas[0].Quantity())
	assert.Equal(t, fmt.Sprintf("%d %s", date.Year(), date.Month().String()), datas[0].Key())

	datas, err = testDB.queryMonth(0, []int{1})
	assert.Nil(t, err)

	assert.Equal(t, 0, len(datas))
}

func TestQueryYear(t *testing.T) {
	datas, err := testDB.queryYear(2, []int{2})
	assert.Nil(t, err)

	assert.Equal(t, 1, len(datas))
	assert.Equal(t, 1200, datas[0].Quantity())
	assert.Equal(t, fmt.Sprintf("%d", date.Year()), datas[0].Key())

	datas, err = testDB.queryYear(0, []int{1})
	assert.Nil(t, err)

	assert.Equal(t, 0, len(datas))
}

func TestQueryLastContext(t *testing.T) {
	data, err := testDB.queryLastRecord(1, DAY)
	assert.Equal(t, "no data", err.Error())

	data, err = testDB.queryLastRecord(2, DAY)
	assert.Nil(t, err)
	assert.Equal(t, 1200, data.Quantity())

	testDB.Close()
}

func TestWithDBContext(t *testing.T) {
	fn := func(db *DB) error { return nil }

	err := withDBContext("", fn)
	assert.Equal(t, err, ErrNoName)

	err = withDBContext("foo", fn)
	assert.Equal(t, err, &ErrInvalidDB{"foo"})

	err = withDBContext("test", fn)
	assert.Nil(t, err)
}
