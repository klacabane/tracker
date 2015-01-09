package main

import (
	"fmt"
	"strconv"
)

const (
	TPL_HEAD = "|  %s"
	TPL_TAIL = "  |\n"
)

type Table struct {
	rows []row
	// row separator
	sep string
	// title of the table
	title string
	// len of the largest key and value ( stringlen )
	// of the rows.
	// Used to build rows with proper alignment
	lenkey, lenval     int
	padding, titlediff int

	isSingle bool
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

func (t *Table) SetSingle(s bool) {
	t.isSingle = s
}

func (t *Table) computeRowSep() {
	for _, row := range t.rows {
		if lenkey := len(row.key); lenkey > t.lenkey {
			t.lenkey = lenkey
		}
		if !t.isSingle {
			if lenval := len(strconv.FormatFloat(row.value, 'f', 2, 64)); lenval > t.lenval {
				t.lenval = lenval
			}
		}
	}

	if len(t.title) > 0 {
		var rowlen int
		if t.isSingle {
			rowlen = t.lenkey
		} else {
			rowlen = t.lenkey + (2 * t.padding) + 1 + t.lenval
		}

		if t.titlediff = rowlen - len(t.title); t.titlediff < 0 {
			// title is larger than the largest row content,
			// add diff to cells
			if t.isSingle {
				t.lenkey -= t.titlediff
			} else {
				complet := -t.titlediff / 2

				t.lenval += complet
				t.lenkey += complet

				if -t.titlediff%2 > 0 {
					t.lenkey++
				}
			}
			t.titlediff = 0
		}
	}

	t.sep = "+"
	for i, j := 0, t.lenkey+t.padding*2; i < j; i++ {
		t.sep += "-"
	}
	t.sep += "+"

	if !t.isSingle {
		for i, j := 0, t.lenval+t.padding*2; i < j; i++ {
			t.sep += "-"
		}
		t.sep += "+"
	}
}

func (t *Table) printRow(r row) {
	var (
		diff   = t.lenkey - len(r.key)
		rowtpl = TPL_HEAD
	)

	if diff > 0 {
		for i := 0; i < diff; i++ {
			rowtpl += " "
		}
	}

	if !t.isSingle {
		rowtpl += "  |  %.2f"

		if diff = t.lenval - len(strconv.FormatFloat(r.value, 'f', 2, 64)); diff > 0 {
			for i := 0; i < diff; i++ {
				rowtpl += " "
			}
		}
	}
	rowtpl += TPL_TAIL

	if t.isSingle {
		fmt.Printf(rowtpl, r.key)
	} else {
		fmt.Printf(rowtpl, r.key, r.value)
	}
	t.printSep()
}

func (t *Table) printTitle() {
	var rowtpl = TPL_HEAD

	if t.titlediff > 0 {
		for i := 0; i < t.titlediff; i++ {
			rowtpl += " "
		}
	}
	rowtpl += TPL_TAIL

	t.printSep()
	fmt.Printf(rowtpl, t.title)
}

func (t *Table) printSep() {
	fmt.Println(t.sep)
}
