package main

import (
	"fmt"
	"sync"
	"time"
)

type Period int

const (
	DAY Period = iota << 1
	WEEK
	MONTH
	YEAR
)

var Periods = map[string]Period{
	"d": DAY,
	"w": WEEK,
	"m": MONTH,
	"y": YEAR,
}

type Fetcher struct {
	frequency  int
	period     Period
	categories []int
	catnames   []string
	trackers   []string
	periodKeys []string
	data       map[string]int64
	resc       chan result
	catnamec   chan string
	quit       chan struct{}
}

func NewFetcher(freq int, period Period, categories []int, trackers []string) *Fetcher {
	return &Fetcher{
		frequency:  freq,
		categories: categories,
		period:     period,
		trackers:   trackers,
		periodKeys: make([]string, freq+1),
		data:       make(map[string]int64),
		resc:       make(chan result, len(trackers)),
		catnamec:   make(chan string, len(categories)),
		quit:       make(chan struct{}, 1),
	}
}

func (f *Fetcher) fetch() {
	defer func() {
		f.quit <- struct{}{}
	}()

	var wg sync.WaitGroup
	wg.Add(len(f.trackers) + 1)

	go f.setKeys(&wg)

	if len(f.categories) == 0 {
		f.catnamec <- "all categories"
	}

	for _, tracker := range f.trackers {
		go func(dbname string) {
			defer wg.Done()

			var res []timeData
			err := withDBContext(dbname, func(db *DB) error {
				var cerr error
				var name string

				for _, category := range f.categories {
					name, cerr = db.getCategory(category)
					if cerr != nil {
						return cerr
					}
					f.catnamec <- name
				}

				switch f.period {
				case DAY:
					res, cerr = db.queryDay(f.frequency, f.categories)
				case WEEK:
					res, cerr = db.queryWeek(f.frequency, f.categories)
				case MONTH:
					res, cerr = db.queryMonth(f.frequency, f.categories)
				case YEAR:
					res, cerr = db.queryYear(f.frequency, f.categories)
				}
				return cerr
			})
			f.resc <- result{err: err, values: res}
		}(tracker)
	}

	wg.Wait()
}

func (f *Fetcher) setKeys(wg *sync.WaitGroup) {
	defer wg.Done()

	tdata := timeData{date: time.Now(), period: f.period}

	for i := f.frequency; i >= 0; i-- {
		f.periodKeys[i] = tdata.Key()
		tdata = tdata.Prev()
	}
}

func (f *Fetcher) Exec() (err error) {
	go f.fetch()

out:
	for {
		select {
		case res := <-f.resc:
			if res.err != nil {
				if err == nil {
					err = res.err
				}
				continue
			}

			for _, data := range res.values {
				f.data[data.Key()] += data.Quantity()
			}
		case name := <-f.catnamec:
			f.catnames = append(f.catnames, name)
		case <-f.quit:
			break out
		}
	}
	return
}

func (f *Fetcher) Data() map[string]float64 {
	rowsf := map[string]float64{}
	for k, v := range f.data {
		rowsf[k] = float64(v) / 100
	}
	return rowsf
}

func (f *Fetcher) PeriodKeys() []string {
	return f.periodKeys
}

func (f *Fetcher) Sum() float64 {
	var sum int64
	for _, v := range f.data {
		sum += v
	}
	return float64(sum) / 100
}

func (f *Fetcher) CatNames() []string {
	return f.catnames
}

type result struct {
	values []timeData
	err    error
}

type timeData struct {
	qty    int64
	date   time.Time
	period Period
}

func (data timeData) Key() string {
	switch data.period {
	case DAY:
		return fmt.Sprintf("%s %02d %s", data.date.Month().String(), data.date.Day(), data.date.Weekday().String())
	case WEEK:
		year, week := data.date.ISOWeek()
		return fmt.Sprintf("W%02d %d", week, year)
	case MONTH:
		return fmt.Sprintf("%d %s", data.date.Year(), data.date.Month().String())
	case YEAR:
		return fmt.Sprintf("%d", data.date.Year())
	}
	return "unknown"
}

func (data timeData) Quantity() int64 {
	return data.qty
}

func (data timeData) Prev() timeData {
	prev := timeData{period: data.period}

	switch data.period {
	case DAY:
		prev.date = data.date.AddDate(0, 0, -1)
	case WEEK:
		prev.date = data.date.AddDate(0, 0, -7)
	case MONTH:
		prev.date = data.date.AddDate(0, -1, 0)
	case YEAR:
		prev.date = data.date.AddDate(-1, 0, 0)
	}
	return prev
}
