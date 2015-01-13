package main

import (
	"fmt"
	"strings"
)

type Table struct {
	Padding int
	Title   string

	rows    [][]string
	columns []*column

	separator string
	titleDiff int
}

type column struct {
	name  string
	width int
}

func NewTableNamedCols(cols ...string) *Table {
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

func NewTable(colNb uint) *Table {
	t := &Table{
		rows:    make([][]string, 0),
		columns: make([]*column, colNb),
		Padding: 2,
	}

	for i := uint(0); i < colNb; i++ {
		t.columns[i] = &column{}
	}
	return t
}

func (t *Table) Print() {
	t.computeSeparator()

	if len(t.Title) > 0 {
		t.printTitle()
	}
	t.printSeparator()

	if t.showColumNames() {
		t.printRow(t.columnNames())
		t.printSeparator()
	}
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

func (t *Table) showColumNames() bool {
	for _, col := range t.columns {
		if len(col.name) > 0 {
			return true
		}
	}
	return false
}

func (t *Table) columnNames() (names []string) {
	for _, col := range t.columns {
		names = append(names, col.name)
	}
	return
}

func (t *Table) adjustTitleDiff() {
	var (
		contentWidth int

		titleWidth = len(t.Title)
		colNb      = len(t.columns)
	)

	if colNb == 1 {
		contentWidth = t.columns[0].width
	} else {
		first, last, border := 0, colNb-1, 1
		for i, col := range t.columns {
			if i == first {
				contentWidth += col.width + t.Padding + border
			} else if i == last {
				contentWidth += col.width + t.Padding
			} else {
				contentWidth += col.width + t.Padding*2 + border
			}
		}
	}

	if diff := titleWidth - contentWidth; diff > 0 {
		chunk := diff / colNb

		for _, col := range t.columns {
			col.width += chunk
		}

		if diff%2 > 0 {
			t.columns[0].width++
		}
	} else {
		t.titleDiff = -diff
	}
}

func (t *Table) computeSeparator() {
	if len(t.Title) > 0 {
		t.adjustTitleDiff()
	}
	t.separator = "+"

	for _, col := range t.columns {
		fullWidth := col.width + t.Padding*2

		t.separator += strings.Repeat("-", fullWidth)
		t.separator += "+"
	}
}

func (t *Table) printRow(row []string) {
	tpl := "|"
	for i, field := range row {
		tpl += strings.Repeat(" ", t.Padding)
		tpl += field

		if diff := t.columns[i].width - len(field); diff > 0 {
			tpl += strings.Repeat(" ", diff)
		}

		tpl += strings.Repeat(" ", t.Padding)
		tpl += "|"
	}

	fmt.Println(tpl)
}

func (t *Table) printTitle() {
	separator := strings.Replace(t.separator[1:len(t.separator)-1], "+", "-", -1)
	separator = "+" + separator + "+"

	diff := t.Padding + t.titleDiff

	fmt.Println(separator)

	tpl := "|"
	tpl += strings.Repeat(" ", t.Padding)
	tpl += strings.ToUpper(t.Title)
	tpl += strings.Repeat(" ", diff)
	tpl += "|"

	fmt.Println(tpl)
}

func (t *Table) printSeparator() {
	fmt.Println(t.separator)
}
