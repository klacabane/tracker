package main

import (
	"bytes"
	"fmt"
	"strings"
)

type Table struct {
	CellPadding int
	Title       string

	rows    [][]string
	columns []*column

	separator string
	titleDiff int
}

type column struct {
	name  string
	width int
}

func NewTableNamedCols(col string, cols ...string) *Table {
	t := &Table{
		rows:        make([][]string, 0),
		columns:     make([]*column, len(cols)+1),
		CellPadding: 2,
	}

	t.columns[0] = &column{col, len(col)}
	for i, c := range cols {
		t.columns[i+1] = &column{c, len(c)}
	}
	return t
}

func NewTable(colNb int) *Table {
	if colNb <= 0 {
		colNb = 1
	}

	t := &Table{
		rows:        make([][]string, 0),
		columns:     make([]*column, colNb),
		CellPadding: 2,
	}

	for i := 0; i < colNb; i++ {
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

	if t.showColumnNames() {
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

func (t *Table) showColumnNames() bool {
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
				contentWidth += col.width + t.CellPadding + border
			} else if i == last {
				contentWidth += col.width + t.CellPadding
			} else {
				contentWidth += col.width + t.CellPadding*2 + border
			}
		}
	}

	if diff := titleWidth - contentWidth; diff <= 0 {
		t.titleDiff = -diff
	} else {
		if colNb == 1 {
			t.columns[0].width += diff
		} else {
			chunk := diff / colNb

			for _, col := range t.columns {
				col.width += chunk
			}

			if diff%2 > 0 {
				t.columns[0].width++
			}
		}
	}
}

func (t *Table) computeSeparator() {
	if len(t.Title) > 0 {
		t.adjustTitleDiff()
	}

	b := bytes.NewBufferString("+")
	for _, col := range t.columns {
		fullWidth := col.width + t.CellPadding*2

		b.WriteString(strings.Repeat("-", fullWidth))
		b.WriteString("+")
	}
	t.separator = b.String()
}

func (t *Table) printRow(row []string) {
	b := bytes.NewBufferString("|")
	for i, field := range row {
		b.WriteString(strings.Repeat(" ", t.CellPadding))
		b.WriteString(field)

		if diff := t.columns[i].width - len(field); diff > 0 {
			b.WriteString(strings.Repeat(" ", diff))
		}

		b.WriteString(strings.Repeat(" ", t.CellPadding))
		b.WriteString("|")
	}

	fmt.Println(b.String())
}

func (t *Table) printTitle() {
	separator := "+" + strings.Replace(t.separator[1:len(t.separator)-1], "+", "-", -1) + "+"

	diff := t.CellPadding + t.titleDiff

	tpl := "|" + strings.Repeat(" ", t.CellPadding) +
		strings.ToUpper(t.Title) + strings.Repeat(" ", diff) + "|"

	fmt.Println(separator)
	fmt.Println(tpl)
}

func (t *Table) printSeparator() {
	fmt.Println(t.separator)
}
