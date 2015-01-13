package main

import (
	"fmt"
	"os"
	"os/user"
	"sort"
	"strings"
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

				table := NewTable("TRACKERS LIST")
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

						table := NewTable("CATEGORIES", "")
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
			},
			Action: func(c *cli.Context) {
				var (
					period    = c.String("period")
					occurence = c.Int("occurence")
					category  = c.Int("category")
					trackers  = c.StringSlice("t")

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

				if len(trackers) == 1 && trackers[0] == "all" {
					var err error

					trackers, err = dblist()
					if err != nil {
						fmt.Println(err)
						return
					}
				}

				if len(trackers) > 1 && category != 0 {
					category = 0
					fmt.Println("ignoring category flag.")
				}

				chres := make(chan result, len(trackers))
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
							title, res.err = db.getCategory(category)
							if res.err != nil {
								chres <- res
								return
							}
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
					total float64

					rows  = make(map[string]float64)
					table = NewTable("", "")
				)
				for i := 0; i < len(trackers); i++ {
					res := <-chres
					if res.err != nil {
						fmt.Println(res.err)
						return
					}

					for _, data := range res.values {
						rows[data.key()] += data.sum()
					}

				}

				for _, k := range <-chkeys {
					sum := rows[k]

					total += sum
					table.Add(k, sum)
				}
				table.Add("Total", total)
				table.SetColumn(0, strings.ToUpper(title))

				table.Print()
			},
		},
		{
			Name: "graph",
			Action: func(c *cli.Context) {
				res := result{
					values: []dataFormatter{
						weekData{22, 2014, 15},
						weekData{52, 2015, 1},
						weekData{200, 2012, 51},
					},
				}
				sort.Sort(res)

				var (
					points = make(map[string]float64)
					labels = make([]string, len(res.values))
				)
				for i, data := range res.values {
					labels[i] = data.key()
					points[data.key()] = data.sum()
				}

				g := NewGraph(labels, points)
				g.Print()
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
