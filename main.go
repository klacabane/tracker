package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	_ "github.com/mattn/go-sqlite3"
)

var (
	TRACKER_DIR, BASE_DB string

	ErrInvalidDB = errors.New("tracker doesnt exist.")
	ErrNoName    = errors.New("tracker name required.")
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
	qty   float64
	month int
}

func (md monthData) sum() float64 {
	return md.qty
}

func (md monthData) key() string {
	return fmt.Sprintf("%d", md.month)
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

				for _, tracker := range trackers {
					fmt.Println(tracker)
				}
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
					var exists bool
					err = db.QueryRow("select 1 from categories where id = ?", category).Scan(&exists)
					if err != nil {
						if err == sql.ErrNoRows {
							fmt.Println("category doesnt exist.")
						} else {
							fmt.Println(err)
						}
						return
					}
				}

				year, week := time.Now().ISOWeek()

				_, err = db.Exec("insert into records(qty, category, isoweek, isoyear) values(?, ?, ?, ?)",
					quantity, category, week, year)
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

						rows, err := db.Query("select id, name from categories")
						if err != nil {
							fmt.Println(err)
							return
						}
						defer rows.Close()

						var (
							id   int64
							name string
						)
						for rows.Next() {
							err = rows.Scan(&id, &name)
							if err != nil {
								fmt.Println(err)
								return
							}
							fmt.Println(id, name)
						}
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

						stmt, err := db.Prepare("insert into categories(name) values(?)")
						if err != nil {
							fmt.Println(err)
							return
						}
						defer stmt.Close()

						for i := 1; i < len(c.Args()); i++ {
							_, err = stmt.Exec(c.Args().Get(i))
							if err != nil {
								fmt.Println(err)
								return
							}
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
			},
			Action: func(c *cli.Context) {
				var (
					period    = c.String("period")
					occurence = c.Int("occurence")
					trackers  = c.StringSlice("t")
				)

				if len(trackers) == 1 && trackers[0] == "all" {
					var err error

					trackers, err = dblist()
					if err != nil {
						fmt.Println(err)
						return
					}
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

						switch period {
						case "w":
							res = queryWeek(db, occurence)
						case "m":
							res = queryMonth(db, occurence)
						case "y":
							res = queryYear(db, occurence)
						default:
							res.err = fmt.Errorf("invalid period argument.")
						}

						ch <- res
					}(tracker)
				}

				var m = make(map[string]float64)

				for i := 0; i < len(trackers); i++ {
					res := <-ch
					if res.err != nil {
						fmt.Println(res.err)
						continue
					}

					for _, data := range res.values {
						m[data.key()] += data.sum()
					}
				}

				for k, v := range m {
					fmt.Println(k, v)
				}
			},
		},
	}

	app.Run(os.Args)
}

func queryWeek(db *sql.DB, occurence int) result {
	var (
		year, week = time.Now().ISOWeek()

		res   result
		query string
	)

	if occurence <= 0 {
		// query concerns current week
		query = fmt.Sprintf("select sum(qty), isoyear, isoweek from records "+
			"where isoyear = %d and isoweek = %d group by isoweek", year, week)
	} else {
		year, week = computeLimitWeek(year, week, occurence)

		query = fmt.Sprintf("select sum(qty), isoyear, isoweek from records "+
			"where (isoyear >= %d and isoweek >= %d) "+
			"or isoyear > %d group by isoweek", year, week, year)
	}

	rows, err := db.Query(query)
	if err != nil {
		res.err = err
		return res
	}
	defer rows.Close()

	var wdata weekData
	for rows.Next() {
		err = rows.Scan(&wdata.qty, &wdata.year, &wdata.week)
		if err != nil {
			res.err = err
			break
		}
		res.values = append(res.values, wdata)
	}

	if res.err == nil {
		res.err = rows.Err()
	}
	return res
}

func queryMonth(db *sql.DB, occurrence int) result {
	return result{}
}

func queryYear(db *sql.DB, occurence int) result {
	return result{}
}

func dblist() ([]string, error) {
	var names []string
	files, err := ioutil.ReadDir(TRACKER_DIR)
	if err != nil {
		return names, err
	}

	for _, f := range files {
		n := f.Name()
		if !f.IsDir() && strings.Contains(n, ".db") {
			names = append(names, strings.TrimSuffix(n, ".db"))
		}
	}

	return names, nil
}

// Helpers
func open(p string) (*sql.DB, error) {
	if !exists(p) {
		create(p)
	}

	db, err := sql.Open("sqlite3", p)
	if err != nil {
		return nil, err
	}
	err = db.Ping()

	return db, err
}

func create(p string) error {
	in, err := os.Open(path.Join(TRACKER_DIR, BASE_DB))
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(p)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func exists(p string) bool {
	if _, err := os.Stat(p); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func openFromContext(c *cli.Context) (*sql.DB, error) {
	dbname := c.Args().First()
	if dbname == "" {
		return nil, ErrNoName
	}

	p := dbpath(dbname)
	if !exists(p) {
		return nil, ErrInvalidDB
	}
	return open(p)
}

func dbpath(name string) string {
	return path.Join(TRACKER_DIR, name+".db")
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
