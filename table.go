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
		if len(k) > t.lenkey {
			t.lenkey = len(k)
		}
		if lenval := len(strconv.Itoa(int(v))) + 3; lenval > t.lenval {
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

func (t *Table) printSep() {
	fmt.Println(t.sep)
}

func (t *Table) printRow(key string, val float64) {
	var (
		dif    = t.lenkey - len(key)
		rowtpl = "|  %s"
	)

	if dif > 0 {
		for i := 0; i < dif; i++ {
			rowtpl += " "
		}
	}
	rowtpl += "  |  %.2f"

	if dif = t.lenval - (len(strconv.Itoa(int(val))) + 3); dif > 0 {
		for i := 0; i < dif; i++ {
			rowtpl += " "
		}
	}
	rowtpl += "  |\n"

	fmt.Printf(rowtpl, key, val)
	t.printSep()
}

func (t *Table) printTitle() {
	var (
		dif    = (len(t.sep) - t.padding*2) - len(t.title)
		rowtpl = "|  %s"
	)

	if dif > 0 {
		for i := 0; i < dif; i++ {
			rowtpl += " "
		}
	} else if dif < 0 {
		// title is larger than the rows,
		// add dif to cells
		t.lenval += -dif / 2
		t.lenkey += -dif / 2

		t.computeRowSep()
	}
	rowtpl += "|\n"

	t.printSep()
	fmt.Printf(rowtpl, t.title)
}
