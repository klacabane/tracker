package main

import (
	"fmt"
)

type Table struct {
	Padding int

	rows      [][]string
	columns   []string
	separator string
	title     string

	columnsSize map[int]int
}

func NewTable(cols ...string) *Table {
	t := &Table{
		rows:        make([][]string, 0),
		columns:     cols,
		columnsSize: make(map[int]int),
		Padding:     2,
	}

	for i, col := range cols {
		t.columnsSize[i] = len(col)
	}
	return t
}

func (t *Table) Print() {
	t.computeSeparator()

	t.printSeparator()
	t.printRow(t.columns)
	t.printSeparator()
	for _, row := range t.rows {
		t.printRow(row)
		t.printSeparator()
	}
}

func (t *Table) Add(row ...interface{}) {
	var (
		lenrow  = len(row)
		lencols = len(t.columns)
	)

	if lenrow == 0 || lenrow > lencols {
		fmt.Println("invalid row len")
		return
	}

	r := make([]string, lencols)
	for i := 0; i < lenrow; i++ {
		val := fmt.Sprintf("%v", row[i])

		if len(val) > t.columnsSize[i] {
			t.columnsSize[i] = len(val)
		}
		r[i] = val
	}
	t.rows = append(t.rows, r)
}

func (t *Table) SetColumn(index int, value string) {
	size, ok := t.columnsSize[index]
	if !ok {
		return
	}

	t.columns[index] = value
	if lencol := len(value); lencol > size {
		t.columnsSize[index] = lencol
	}
}

func (t *Table) computeSeparator() {
	t.separator = "+"

	for _, size := range t.columnsSize {
		fullsize := size + t.Padding*2
		for j := 0; j < fullsize; j++ {
			t.separator += "-"
		}
		t.separator += "+"
	}
}

func (t *Table) printRow(row []string) {
	tpl := "|"
	for i, field := range row {
		for j := 0; j < t.Padding; j++ {
			tpl += " "
		}
		tpl += field

		if diff := t.columnsSize[i] - len(field); diff > 0 {
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
