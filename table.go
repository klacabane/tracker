package main

import (
	"fmt"
	"strconv"
)

type Table struct {
	Data map[string]float64

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

func NewTable() *Table {
	return &Table{
		Data:    make(map[string]float64),
		padding: 2,
	}
}

func (t *Table) Print() {
	t.computeRowSep()
	if len(t.title) > 0 {
		t.printTitle()
	}
	t.printSep()

	var total float64
	for k, v := range t.Data {
		total += v
		t.printRow(k, v)
	}
	t.printRow("Total", total)
}

func (t *Table) SetTitle(title string) {
	t.title = title
}

func (t *Table) computeRowSep() {
	for k, v := range t.Data {
		if lenkey := len(k); lenkey > t.lenkey {
			t.lenkey = lenkey
		}
		if lenval := len(strconv.FormatFloat(v, 'f', 2, 64)); lenval > t.lenval {
			t.lenval = lenval
		}
	}

	if t.lenkey < 5 {
		t.lenkey = 5 // len("Total")
	}
	if t.lenval < 4 {
		t.lenval = 4 // len("0.00")
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

func (t *Table) printRow(key string, val float64) {
	var (
		diff   = t.lenkey - len(key)
		rowtpl = "|  %s"
	)

	if diff > 0 {
		for i := 0; i < diff; i++ {
			rowtpl += " "
		}
	}
	rowtpl += "  |  %.2f"

	if diff = t.lenval - (len(strconv.Itoa(int(val))) + 3); diff > 0 {
		for i := 0; i < diff; i++ {
			rowtpl += " "
		}
	}
	rowtpl += "  |\n"

	fmt.Printf(rowtpl, key, val)
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
