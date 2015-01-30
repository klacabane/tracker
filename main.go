package main

import (
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/codegangsta/cli"
)

var (
	TRACKER_DIR, BASE_DB, DEFAULT_DB string
)

func init() {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	TRACKER_DIR = u.HomeDir + "/Dropbox/tracker/"
	BASE_DB = "conf/base"
	DEFAULT_DB = "default"
}

func main() {
	app := cli.NewApp()
	app.Name = "tracker"
	app.Commands = []cli.Command{
		// List
		{
			Name:  "list",
			Usage: "Lists the existing trackers",
			Action: func(c *cli.Context) {
				trackers, err := dblist()
				if err != nil {
					fmt.Println(err)
					return
				}

				table := NewTable(1)
				table.Title = "TRACKERS"
				for _, tracker := range trackers {
					table.Add(tracker)
				}
				table.Print()
			},
		},
		// New
		{
			Name:  "new",
			Usage: "Creates a new tracker",
			Action: func(c *cli.Context) {
				dbname := c.Args().First()
				if dbname == "" {
					fmt.Println(ErrNoName)
					return
				}

				p := dbpath(dbname)
				if exists(p) {
					fmt.Println("tracker already exists.")
					return
				}

				err := create(p)
				if err != nil {
					fmt.Println(err)
				}
			},
		},
		// Add
		{
			Name: "add",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "tracker, t",
					Value: DEFAULT_DB,
				},
				cli.Float64Flag{
					Name:  "quantity, qty",
					Value: 0,
				},
				cli.IntFlag{
					Name:  "category, cat",
					Value: 1,
				},
			},
			Action: func(c *cli.Context) {
				var (
					category int
					quantity int64
				)

				if qtyf := c.Float64("qty"); qtyf == 0 {
					fmt.Println("no quantity specified.")
					return
				} else {
					quantity = int64(qtyf * 100)
				}

				if err := withDBContext(c.String("t"), func(db *DB) error {
					if category = c.Int("cat"); category != 1 {
						if _, cerr := db.getCategory(category); cerr != nil {
							return cerr
						}
					}
					year, week := time.Now().ISOWeek()

					return db.addRecord(quantity, category, year, week)
				}); err != nil {
					fmt.Println(err)
				}
			},
		},
		// Category
		{
			Name:      "category",
			ShortName: "cat",
			Subcommands: []cli.Command{
				{
					Name: "list",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "tracker, t",
							Value: DEFAULT_DB,
						},
					},
					Action: func(c *cli.Context) {
						if err := withDBContext(c.String("t"), func(db *DB) error {
							categories, cerr := db.getCategories()
							if cerr != nil {
								return cerr
							}

							table := NewTable(2)
							table.Title = "CATEGORIES"
							for k, v := range categories {
								table.Add(k, v)
							}
							table.Print()

							return nil
						}); err != nil {
							fmt.Println(err)
						}
					},
				},
				{
					Name: "add",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "tracker, t",
							Value: DEFAULT_DB,
						},
						cli.StringSliceFlag{
							Name:  "categories, cat",
							Value: &cli.StringSlice{},
						},
					},
					Action: func(c *cli.Context) {
						if length := len(c.StringSlice("cat")); length == 0 {
							fmt.Println("no categories specified.")
							return
						}

						if err := withDBContext(c.String("t"), func(db *DB) error {
							return db.addCategories(c.StringSlice("cat")...)
						}); err != nil {
							fmt.Println(err)
						}
					},
				},
			},
		},
		// aggregate
		{
			Name:      "aggregate",
			ShortName: "agg",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "trackers, t",
					Value: &cli.StringSlice{},
				},
				cli.StringFlag{
					Name:  "period, p",
					Value: "w",
				},
				cli.IntFlag{
					Name:  "occurence, o",
					Value: 2,
				},
				cli.IntFlag{
					Name:  "category, cat",
					Value: 0,
				},
				cli.BoolFlag{
					Name: "graph, g",
				},
			},
			Action: func(c *cli.Context) {
				var (
					period      = c.String("period")
					occurence   = c.Int("occurence")
					category    = c.Int("category")
					trackers    = c.StringSlice("t")
					trackerslen = len(trackers)

					chkeys = make(chan []string, 1)
				)

				if occurence < 0 {
					occurence = 0
				}

				title, err := periodLabel(period)
				if err != nil {
					fmt.Println(err)
					return
				}

				go keys(period, occurence, chkeys)

				if trackerslen == 0 {
					trackers = append(trackers, DEFAULT_DB)
					trackerslen++
				}

				if trackerslen == 1 {
					if title = trackers[0]; title == "all" {
						var err error

						trackers, err = dblist()
						if err != nil {
							fmt.Println(err)
							return
						}
						trackerslen = len(trackers)
					}
				}

				if trackerslen > 1 && category != 0 {
					category = 0
					fmt.Println("ignoring category flag.")
				}

				chres := make(chan result, trackerslen)
				for _, tracker := range trackers {
					go func(dbname string) {
						var res []timeData

						err = withDBContext(dbname, func(db *DB) error {
							var cerr error

							if category > 0 {
								name, cerr := db.getCategory(category)
								if cerr != nil {
									return cerr
								}
								title += " " + name
							}

							switch period {
							case "w":
								res, cerr = db.queryWeek(occurence, category)
							case "m":
								res, cerr = db.queryMonth(occurence, category)
							case "y":
								res, cerr = db.queryYear(occurence, category)
							}
							return cerr
						})
						chres <- result{err: err, values: res}
					}(tracker)
				}

				var (
					component UIComponent

					rows = make(map[string]int64)
				)
				for i := 0; i < trackerslen; i++ {
					res := <-chres
					if res.err != nil {
						fmt.Println(res.err)
						if res.err == ErrInvalidDB {
							continue
						}
						return
					}

					for _, data := range res.values {
						rows[data.Key()] += data.Quantity()
					}
				}

				labels := <-chkeys
				if c.Bool("graph") {
					rowsf := make(map[string]float64)
					for k, v := range rows {
						rowsf[k] = float64(v) / 100
					}

					component = NewGraph(labels, rowsf)
				} else {
					table, total := NewTable(2), int64(0)
					for _, k := range labels {
						sum := rows[k]

						total += sum
						table.Add(k, float64(sum)/100)
					}
					table.Add("Total", float64(total)/100)
					table.Title = title

					component = table
				}
				component.Print()
			},
		},
		{
			Name: "graph",
			Action: func(c *cli.Context) {
				m := map[string]float64{
					"1": 0.15,
					"2": 0.100,
					"3": 0.3000,
					"4": 0.700,
					"5": 0.1500,
					"6": 0.110,
					"7": 0.2900,
					"8": 0.35551,
				}
				graph := NewGraph([]string{"1", "2", "3", "4", "5", "6", "7", "8"}, m)
				graph.Print()
			},
		},
	}

	app.Run(os.Args)
}

func computeLimitMonth(year, month, n int) (int, int) {
	for n > 0 {
		month--
		if month == 0 {
			year--
			month = 12
		}
		n--
	}
	return year, month
}

func computeLimitWeek(year, week, n int) (int, int) {
	for n > 0 {
		week--
		if week == 0 {
			year--
			if isLongYear(year) {
				week = 53
			} else {
				week = 52
			}
		}
		n--
	}
	return year, week
}

func keys(p string, n int, c chan<- []string) {
	var (
		keys = make([]string, n+1)

		date       = time.Now()
		year, week = date.ISOWeek()
		month      = int(date.Month())

		data timeData
	)

	if p == "w" {
		data = timeData{year: year, week: week}
	} else if p == "m" {
		data = timeData{year: year, month: month}
	} else {
		data = timeData{year: year}
	}

	for n >= 0 {
		keys[n] = data.Key()
		data = data.Prev()

		n--
	}
	c <- keys
}

func isLongYear(year int) bool {
	var (
		start = time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
		end   = time.Date(year, time.December, 31, 0, 0, 0, 0, time.Local)

		isLeap = year%4 == 0 && year%100 != 0 || year%400 == 0
	)
	return (isLeap && (start.Weekday() == time.Wednesday || end.Weekday() == time.Friday)) ||
		(!isLeap && (start.Weekday() == time.Thursday || end.Weekday() == time.Friday))
}

func periodLabel(p string) (string, error) {
	switch p {
	case "w":
		return "week", nil
	case "m":
		return "month", nil
	case "y":
		return "year", nil
	default:
		return "", fmt.Errorf("invalid period flag.")
	}
}
