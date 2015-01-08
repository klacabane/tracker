package main

import (
	"fmt"
)

type dataPrinter interface {
	sum() float64
	key() string
}

type result struct {
	values []dataPrinter
	err    error
}

type yearData struct {
	qty  float64
	year int
}

func (yd yearData) sum() float64 {
	return yd.qty
}

func (yd yearData) key() string {
	return fmt.Sprintf("%d", yd.year)
}

type monthData struct {
	qty         float64
	year, month int
}

func (md monthData) sum() float64 {
	return md.qty
}

func (md monthData) key() string {
	return fmt.Sprintf("%d-%s", md.year, monthVal(md.month))
}

type weekData struct {
	qty        float64
	year, week int
}

func (wd weekData) sum() float64 {
	return wd.qty
}

func (wd weekData) key() string {
	return fmt.Sprintf("%d-W%d", wd.year, wd.week)
}
