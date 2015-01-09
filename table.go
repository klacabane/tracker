package main

import "fmt"

type Table struct {
	Data  map[string]float64
	sep   string
	title string
}

func NewTable() *Table {
	return &Table{
		Data: make(map[string]float64),
	}
}

func (t *Table) Print() {
	t.computeRowSep()
	t.printSep()

	if len(t.title) > 0 {
		t.printRow(t.title)
	}

	var total float64
	for k, v := range t.Data {
		total += v
		t.printRow(rowToString(k, v))
	}
	t.printRow(fmt.Sprintf("Total: %.2f", total))
}

func (t *Table) SetTitle(title string) {
	t.title = title
}

func (t *Table) computeRowSep() {
	var longest int
	for k, v := range t.Data {
		length := len(rowToString(k, v))

		if length > longest {
			longest = length
		}
	}

	if longest == 0 {
		longest = 11 // len("Total: 0.00")
	}

	for i := 0; i < longest; i++ {
		t.sep += "-"
	}
}

func (t *Table) printSep() {
	fmt.Println(t.sep)
}

func (t *Table) printRow(row string) {
	fmt.Println(row)
	t.printSep()
}

func rowToString(key string, val float64) string {
	return fmt.Sprintf("%s  |  %.2f", key, val)
}
