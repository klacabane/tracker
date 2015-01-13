package main

import (
	"fmt"
)

type Table struct {
	Padding int

	rows    [][]string
	columns []*column

	separator string
	title     string
}

type column struct {
	name  string
	width int
}

func NewTable(cols ...string) *Table {
	t := &Table{
		rows:    make([][]string, 0),
		columns: make([]*column, len(cols)),
		Padding: 2,
	}

	for i, col := range cols {
		c := &column{col, len(col)}

		t.columns[i] = c
	}
	return t
}

func (t *Table) Print() {
	t.computeSeparator()

	t.printSeparator()
	t.printRow(t.columnNames())
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

		if width := len(val); width > t.columns[i].width {
			t.columns[i].width = width
		}
		r[i] = val
	}
	t.rows = append(t.rows, r)
}

func (t *Table) SetColumn(index int, value string) {
	if index > len(t.columns)-1 {
		return
	}

	col := t.columns[index]

	col.name = value
	if width := len(value); width > col.width {
		col.width = width
	}
}

func (t *Table) columnNames() (names []string) {
	for _, col := range t.columns {
		names = append(names, col.name)
	}
	return
}

func (t *Table) computeSeparator() {
	t.separator = "+"

	for _, col := range t.columns {
		fullsize := col.width + t.Padding*2
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

		if diff := t.columns[i].width - len(field); diff > 0 {
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
