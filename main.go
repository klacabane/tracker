package main

import (
	"fmt"
	"os"
	"os/user"

	"github.com/codegangsta/cli"
)

var (
	TRACKER_DIR, DEFAULT_DB string
)

func init() {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}

	TRACKER_DIR = u.HomeDir + "/Dropbox/tracker/"
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
					printErr(err)
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
					printErr(err)
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

					return db.addRecord(quantity, category)
				}); err != nil {
					printErr(err)
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
							printErr(err)
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
							printErr(err)
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
					Name:  "frequency, f",
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
					period      = Periods[c.String("period")]
					frequency   = c.Int("frequency")
					category    = c.Int("category")
					trackers    = c.StringSlice("t")
					trackerslen = len(trackers)
				)

				if frequency < 0 {
					frequency = 0
				}

				if trackerslen == 0 {
					trackers = append(trackers, DEFAULT_DB)
					trackerslen++
				}

				if trackerslen == 1 && trackers[0] == "all" {
					var err error

					trackers, err = dblist()
					if err != nil {
						printErr(err)
						return
					}
					trackerslen = len(trackers)
				}

				if trackerslen > 1 && category != 0 {
					category = 0
					fmt.Println("ignoring category flag.")
				}

				f := NewFetcher(frequency, category,
					period, trackers)
				if err := f.Exec(); err != nil {
					printErr(err)
					return
				}

				var component UIComponent
				if c.Bool("graph") {
					component = NewGraph(f.Keys(), f.Data())
				} else {
					table, data := NewTable(2), f.Data()
					for _, k := range f.Keys() {
						table.Add(k, data[k])
					}
					table.Add("Total", f.Sum())
					table.Title = f.Title()

					component = table
				}
				component.Print()
			},
		},
	}

	app.Run(os.Args)
}

func printErr(err error) {
	fmt.Println("ERROR:", err.Error())
}
