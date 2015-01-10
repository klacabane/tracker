package main

import (
	"fmt"
)

type Table struct {
	Columns []string
	Padding int

	rows      [][]string
	separator string
	title     string

	columnsSize map[int]int
}

func NewTable(cols ...string) *Table {
	t := &Table{
		rows:        make([][]string, 0),
		Columns:     cols,
		columnsSize: make(map[int]int),
		Padding:     2,
	}

	for i := 0; i < len(cols); i++ {
		t.columnsSize[i] = len(cols[i])
	}
	return t
}

func (t *Table) Print() {
	t.computeSeparator()

	t.printSeparator()
	t.printRow(t.Columns)
	t.printSeparator()
	for i := 0; i < len(t.rows); i++ {
		t.printRow(t.rows[i])
		t.printSeparator()
	}
}

func (t *Table) Add(row ...interface{}) {
	if len(row) == 0 || len(row) != len(t.Columns) {
		fmt.Println("invalid row len")
		return
	}

	r := make([]string, len(t.Columns))
	for i := 0; i < len(t.Columns); i++ {
		val := fmt.Sprintf("%v", row[i])

		if len(val) > t.columnsSize[i] {
			t.columnsSize[i] = len(val)
		}
		r[i] = val
	}
	t.rows = append(t.rows, r)
}

func (t *Table) computeSeparator() {
	t.separator = "+"

	for i := 0; i < len(t.columnsSize); i++ {
		colsize := t.columnsSize[i] + t.Padding*2
		for j := 0; j < colsize; j++ {
			t.separator += "-"
		}
		t.separator += "+"
	}
}

func (t *Table) printRow(row []string) {
	tpl := "|"
	for i := 0; i < len(row); i++ {
		val := row[i]

		for j := 0; j < t.Padding; j++ {
			tpl += " "
		}
		tpl += val

		if diff := t.columnsSize[i] - len(val); diff > 0 {
			for j := 0; j < diff; j++ {
				tpl += " "
			}
		}

		for j := 0; j < t.Padding; j++ {
			tpl += " "
		}
		tpl += "|"
	}

	fmt.Println(tpl)
}

func (t *Table) printSeparator() {
	fmt.Println(t.separator)
}
