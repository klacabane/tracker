package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTable(t *testing.T) {
	title := "TEST VERY LONG TITLE"
	row := row{"R", 12.12}

	table := NewTable()

	table.SetTitle(title)
	table.Append(row.key, row.value)

	table.computeRowSep()
	assert.Equal(t, 0, table.titlediff, "Title diff should be 0")
	assert.Equal(t, len(title), table.lenkey+(2*table.padding)+1+table.lenval, "Row len should be equal to title len")

	table.lenkey, table.lenval = 0, 0

	table.SetTitle("")
	table.computeRowSep()
	assert.Equal(t, 1, table.lenkey)
	assert.Equal(t, 5, table.lenval)

	assert.Equal(t, "+-----+---------+", table.sep)
}

func TestSingleTable(t *testing.T) {
	title := "TEST TITLE"
	table := NewTable()

	table.SetTitle(title)
	table.SetSingle(true)
	table.computeRowSep()

	assert.Equal(t, len(title)+(2*table.padding)+2, len(table.sep))

	key := "SUPER LONG KEY"
	table.Append(key, 0)
	table.computeRowSep()
	assert.Equal(t, len(key)-len(title), table.titlediff)
	assert.Equal(t, "+------------------+", table.sep)
}
