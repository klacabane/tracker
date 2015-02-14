package main

import (
	"fmt"
	"strings"
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
	title          string
	freq, category int
	period         Period
	trackers       []string
	sortedKeys     []string
	data           map[string]int64
	resc           chan result
}

func NewFetcher(freq, category int, period Period, trackers []string) *Fetcher {
	return &Fetcher{
		title:      strings.Join(trackers, " & "),
		freq:       freq,
		category:   category,
		period:     period,
		trackers:   trackers,
		sortedKeys: make([]string, freq+1),
		data:       make(map[string]int64),
		resc:       make(chan result, len(trackers)),
	}
}

func (f *Fetcher) fetch() {
	var wg sync.WaitGroup
	wg.Add(len(f.trackers) + 1)

	go func() {
		defer wg.Done()
		f.setKeys()
	}()

	for _, tracker := range f.trackers {
		go func(dbname string) {
			defer wg.Done()

			var res []timeData
			err := withDBContext(dbname, func(db *DB) error {
				var cerr error

				if f.category > 0 {
					if f.title, cerr = db.getCategory(f.category); cerr != nil {
						return cerr
					}
				}

				switch f.period {
				case DAY:
					res, cerr = db.queryDay(f.freq, f.category)
				case WEEK:
					res, cerr = db.queryWeek(f.freq, f.category)
				case MONTH:
					res, cerr = db.queryMonth(f.freq, f.category)
				case YEAR:
					res, cerr = db.queryYear(f.freq, f.category)
				}
				return cerr
			})
			f.resc <- result{err: err, values: res}
		}(tracker)
	}

	wg.Wait()
	close(f.resc)
}

func (f *Fetcher) setKeys() {
	tdata := timeData{date: time.Now(), period: f.period}

	for i := f.freq; i >= 0; i-- {
		f.sortedKeys[i] = tdata.Key()
		tdata = tdata.Prev()
	}
}

func (f *Fetcher) Exec() (err error) {
	go f.fetch()

	for res := range f.resc {
		if res.err != nil {
			if err == nil {
				err = res.err
			}
			continue
		}

		for _, data := range res.values {
			f.data[data.Key()] += data.Quantity()
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

func (f *Fetcher) Keys() []string {
	return f.sortedKeys
}

func (f *Fetcher) Sum() float64 {
	var sum int64
	for _, v := range f.data {
		sum += v
	}
	return float64(sum) / 100
}

func (f *Fetcher) Title() string {
	return f.title
}

type UIComponent interface {
	Print()
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
