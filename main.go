package main

import (
	"fmt"
	"os"
	"os/user"
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

				table := NewTable()
				table.SetTitle("TRACKERS LIST")
				table.SetSingle(true)

				for _, tracker := range trackers {
					table.Append(tracker, 0)
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
						table := NewTable()
						table.SetTitle("CATEGORIES")
						table.SetSingle(true)

						for k, v := range categories {
							table.Append(v, float64(k))
						}
						table.Print()
					},
				},
				{
					Name: "add",
					Action: func(c *cli.Context) {
						if len(c.Args()) <= 1 {
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
					table = NewTable()

					period    = c.String("period")
					occurence = c.Int("occurence")
					category  = c.Int("category")
					trackers  = c.StringSlice("t")
				)

				if period != "w" && period != "m" && period != "y" {
					fmt.Println("invalid period flag.")
					return
				}

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

				ch := make(chan result, len(trackers))
				for _, tracker := range trackers {
					go func(dbname string) {
						var res result

						p := dbpath(dbname)
						if !exists(p) {
							res.err = ErrInvalidDB
							ch <- res
							return
						}

						db, err := open(p)
						if err != nil {
							res.err = err
							ch <- res
							return
						}
						defer db.Close()

						if category != 0 {
							res.title, res.err = db.getCategory(category)
							if res.err != nil {
								ch <- res
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

						ch <- res
					}(tracker)
				}

				var rows = make(map[string]float64)
				for i := 0; i < len(trackers); i++ {
					res := <-ch
					if res.err != nil {
						fmt.Println(res.err)
						continue
					}

					for _, data := range res.values {
						rows[data.key()] += data.sum()
					}

					if len(res.title) > 0 {
						table.SetTitle(strings.ToUpper(res.title))
					}
				}

				var total float64
				for k, v := range rows {
					total += v
					table.Append(k, v)
				}
				table.Append("Total", total)

				table.Print()
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

func isLongYear(year int) bool {
	var (
		start = time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
		end   = time.Date(year, time.December, 31, 0, 0, 0, 0, time.Local)

		isLeap = year%4 == 0 && year%100 != 0 || year%400 == 0
	)
	return (isLeap && (start.Weekday() == time.Wednesday || end.Weekday() == time.Friday)) ||
		(!isLeap && (start.Weekday() == time.Thursday || end.Weekday() == time.Friday))
}
