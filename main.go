package main

import (
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/codegangsta/cli"
)

var (
	TRACKER_DIR, BASE_DB string
)

func init() {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	TRACKER_DIR = u.HomeDir + "/Dropbox/tracker/"
	BASE_DB = "conf/base.db"
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
					fmt.Println("name required.")
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
					quantity float64
				)

				if quantity = c.Float64("qty"); quantity == 0 {
					fmt.Println("no quantity specified.")
					return
				}

				db, err := openFromContext(c)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer db.Close()

				if category = c.Int("cat"); category != 1 {
					_, err = db.getCategory(category)
					if err != nil {
						fmt.Println(err)
						return
					}
				}

				year, week := time.Now().ISOWeek()

				err = db.addRecord(quantity, category, year, week)
				if err != nil {
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
					Action: func(c *cli.Context) {
						db, err := openFromContext(c)
						if err != nil {
							fmt.Println(err)
							return
						}
						defer db.Close()

						categories, err := db.getCategories()
						if err != nil {
							fmt.Println(err)
							return
						}

						table := NewTable(2)
						table.Title = "CATEGORIES"
						for k, v := range categories {
							table.Add(k, v)
						}
						table.Print()
					},
				},
				{
					Name: "add",
					Action: func(c *cli.Context) {
						if length := len(c.Args()); length == 0 {
							fmt.Println(ErrNoName)
							return
						} else if length == 1 {
							fmt.Println("no categories specified.")
							return
						}

						db, err := openFromContext(c)
						if err != nil {
							fmt.Println(err)
							return
						}
						defer db.Close()

						err = db.addCategories(c.Args()[1:]...)
						if err != nil {
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
					Name:  "tracker, t",
					Value: &cli.StringSlice{},
				},
				cli.StringFlag{
					Name:  "period, p",
					Value: "w",
				},
				cli.IntFlag{
					Name:  "occurence, o",
					Value: 0,
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
						var res result

						p := dbpath(dbname)
						if !exists(p) {
							res.err = ErrInvalidDB
							chres <- res
							return
						}

						db, err := open(p)
						if err != nil {
							res.err = err
							chres <- res
							return
						}
						defer db.Close()

						if category > 0 {
							name, err := db.getCategory(category)
							if err != nil {
								res.err = err
								chres <- res
								return
							}
							title += " " + name
						}

						switch period {
						case "w":
							res.values, res.err = db.queryWeek(occurence, category)
						case "m":
							res.values, res.err = db.queryMonth(occurence, category)
						case "y":
							res.values, res.err = db.queryYear(occurence, category)
						}

						chres <- res
					}(tracker)
				}

				var (
					component UIComponent

					rows = make(map[string]float64)
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
						rows[data.key()] += data.sum()
					}

				}

				labels := <-chkeys
				if c.Bool("graph") {
					component = NewGraph(labels, rows)
				} else {
					table, total := NewTable(2), float64(0)
					for _, k := range labels {
						sum := rows[k]

						total += sum
						table.Add(k, sum)
					}
					table.Add("Total", total)
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
					"1": 15,
					"2": 100,
					"3": 3000,
					"4": 700,
					"5": 1500,
					"6": 110,
					"7": 2900,
					"8": 3500,
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
	if p == "w" {
		c <- weekKeys(n)
	} else if p == "m" {
		c <- monthKeys(n)
	} else {
		c <- yearKeys(n)
	}
}

func weekKeys(n int) []string {
	var (
		keys       = make([]string, n+1)
		year, week = time.Now().ISOWeek()

		wd = weekData{0, year, week}
	)

	for n >= 0 {
		keys[n] = wd.key()
		wd.year, wd.week = computeLimitWeek(wd.year, wd.week, 1)

		n--
	}
	return keys
}

func monthKeys(n int) []string {
	var (
		keys = make([]string, n+1)
		date = time.Now()

		md = monthData{0, date.Year(), int(date.Month())}
	)

	for n >= 0 {
		keys[n] = md.key()
		md.year, md.month = computeLimitMonth(md.year, md.month, 1)

		n--
	}
	return keys
}

func yearKeys(n int) []string {
	var (
		keys = make([]string, n+1)

		yd = yearData{0, time.Now().Year()}
	)

	for n >= 0 {
		keys[n] = yd.key()
		yd.year--

		n--
	}
	return keys
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
