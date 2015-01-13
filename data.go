package main

import (
	"fmt"
)

type dataFormatter interface {
	sum() float64
	key() string
	less(dataFormatter) bool
}

type result struct {
	values []dataFormatter
	err    error
}

func (r result) Len() int {
	return len(r.values)
}

func (r result) Swap(i, j int) {
	r.values[i], r.values[j] = r.values[j], r.values[i]
}

func (r result) Less(i, j int) bool {
	return r.values[i].less(r.values[j])
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

func (yd yearData) less(comp dataFormatter) bool {
	return yd.year < comp.(yearData).year
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

func (md monthData) less(v dataFormatter) bool {
	comp := v.(monthData)

	return md.year < comp.year || (md.year == comp.year && md.month < comp.month)
}

type weekData struct {
	qty        float64
	year, week int
}

func (wd weekData) sum() float64 {
	return wd.qty
}

func (wd weekData) key() string {
	return fmt.Sprintf("%d-W%s", wd.year, monthVal(wd.week))
}

func (wd weekData) less(v dataFormatter) bool {
	comp := v.(weekData)

	return wd.year < comp.year || (wd.year == comp.year && wd.week < comp.week)

}
