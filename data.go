package main

import (
	"fmt"
	"strings"
	"sync"
)

type Fetcher struct {
	title, period  string
	freq, category int
	trackers       []string
	data           map[string]int64
	resc           chan result
}

func NewFetcher(freq, category int, period string, trackers []string) *Fetcher {
	return &Fetcher{
		period:   period,
		freq:     freq,
		category: category,
		trackers: trackers,
		resc:     make(chan result, len(trackers)),
		data:     make(map[string]int64),
		title:    strings.Join(trackers, " & "),
	}
}

func (f *Fetcher) fetch() {
	var wg sync.WaitGroup
	wg.Add(len(f.trackers))

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
				case "w":
					res, cerr = db.queryWeek(f.freq, f.category)
				case "m":
					res, cerr = db.queryMonth(f.freq, f.category)
				case "y":
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
		return fmt.Sprintf("%d-W%02d", data.year, data.week)
	} else if data.month > 0 {
		return fmt.Sprintf("%d-%02d", data.year, data.month)
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
