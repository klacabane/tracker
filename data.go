package main

import (
	"fmt"
)

type UIComponent interface {
	Print()
}

type result struct {
	values []timeData
	err    error
}

type timeData struct {
	qty               int64
	year, month, week int
}

func (data timeData) Key() string {
	if data.week > 0 {
		return fmt.Sprintf("%d-W%s", data.year, monthVal(data.week))
	} else if data.month > 0 {
		return fmt.Sprintf("%d-%s", data.year, monthVal(data.month))
	}
	return fmt.Sprintf("%d", data.year)
}

func (data timeData) Quantity() int64 {
	return data.qty
}

func (data timeData) Prev() timeData {
	var prev timeData

	if data.week > 0 {
		prev.year, prev.week = computeLimitWeek(data.year, data.week, 1)
	} else if data.month > 0 {
		prev.year, prev.month = computeLimitMonth(data.year, data.month, 1)
	} else {
		prev.year = data.year - 1
	}
	return prev
}
