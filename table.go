package main

import (
	"fmt"
	"strconv"
)

type Table struct {
	rows []row
	// row separator
	sep string
	// title of the table
	title string
	// len of the largest key and value ( stringlen )
	// of the Data map.
	// Used to build rows with proper alignment
	lenkey, lenval int

	padding int
}

type row struct {
	key   string
	value float64
}

func NewTable() *Table {
	return &Table{
		rows:    make([]row, 0),
		padding: 2,
	}
}

func (t *Table) Print() {
	t.computeRowSep()
	if len(t.title) > 0 {
		t.printTitle()
	}
	t.printSep()

	for _, row := range t.rows {
		t.printRow(row)
	}
}

func (t *Table) Append(key string, value float64) {
	t.rows = append(t.rows, row{key, value})
}

func (t *Table) SetTitle(title string) {
	t.title = title
}

func (t *Table) computeRowSep() {
	for _, row := range t.rows {
		if lenkey := len(row.key); lenkey > t.lenkey {
			t.lenkey = lenkey
		}
		if lenval := len(strconv.FormatFloat(row.value, 'f', 2, 64)); lenval > t.lenval {
			t.lenval = lenval
		}
	}

	t.sep = "+"
	for i := 0; i < t.lenkey+t.padding*2; i++ {
		t.sep += "-"
	}
	t.sep += "+"
	for i := 0; i < t.lenval+t.padding*2; i++ {
		t.sep += "-"
	}
	t.sep += "+"
}

func (t *Table) printRow(r row) {
	var (
		diff   = t.lenkey - len(r.key)
		rowtpl = "|  %s"
	)

	if diff > 0 {
		for i := 0; i < diff; i++ {
			rowtpl += " "
		}
	}
	rowtpl += "  |  %.2f"

	if diff = t.lenval - len(strconv.FormatFloat(r.value, 'f', 2, 64)); diff > 0 {
		for i := 0; i < diff; i++ {
			rowtpl += " "
		}
	}
	rowtpl += "  |\n"

	fmt.Printf(rowtpl, r.key, r.value)
	t.printSep()
}

func (t *Table) printTitle() {
	var (
		diff   = (len(t.sep) - (t.padding*2 + 2)) /*padding+borders*/ - len(t.title)
		rowtpl = "|  %s"
	)

	if diff > 0 {
		for i := 0; i < diff; i++ {
			rowtpl += " "
		}
	} else if diff < 0 {
		complet := -diff / 2
		// title is larger than the rows,
		// add diff to cells
		t.lenval += complet
		t.lenkey += complet

		if -diff%2 > 0 {
			t.lenkey++
		}

		t.computeRowSep()
	}
	rowtpl += "  |\n"

	t.printSep()
	fmt.Printf(rowtpl, t.title)
}

func (t *Table) printSep() {
	fmt.Println(t.sep)
}
