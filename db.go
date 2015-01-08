package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrInvalidDB       = errors.New("tracker doesnt exist.")
	ErrInvalidCategory = errors.New("category doesnt exist.")
	ErrNoName          = errors.New("tracker name required.")
)

type DB struct {
	*sql.DB
}

func (db *DB) queryWeek(occurence int) result {
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
			return res
		}
		res.values = append(res.values, wdata)
	}

	res.err = rows.Err()
	return res
}

func (db *DB) queryMonth(occurence int) result {
	var (
		date        = time.Now()
		year, month = strconv.Itoa(date.Year()), monthVal(int(date.Month()))

		res   result
		query string
	)

	if occurence <= 0 {
		query = "select sum(qty), isoyear, strftime('%m', date) " +
			"from records where strftime('%m', date) = " + fmt.Sprintf("'%s' ", month) +
			"and strftime('%Y', date) = " + fmt.Sprintf("'%s'", year)
	} else {
		query = "select sum(qty), isoyear, strftime('%m', date) " +
			"from records group by strftime('%Y-%m', date)"
	}

	rows, err := db.Query(query)
	if err != nil {
		res.err = err
		return res
	}
	defer rows.Close()

	var mdata monthData
	for rows.Next() {
		err = rows.Scan(&mdata.qty, &mdata.year, &mdata.month)
		if err != nil {
			res.err = err
			return res
		}
		res.values = append(res.values, mdata)
	}

	res.err = rows.Err()
	return res
}

func (db *DB) queryYear(occurence int) result {
	var (
		year = time.Now().Year()

		res   result
		query string
	)

	if occurence <= 0 {
		query = fmt.Sprintf("select sum(qty), isoyear from records "+
			"where isoyear = %d", year)
	} else {
		year = year - occurence

		query = fmt.Sprintf("select sum(qty), isoyear from records "+
			"where isoyear >= %d group by isoyear", year)
	}

	rows, err := db.Query(query)
	if err != nil {
		res.err = err
		return res
	}
	defer rows.Close()

	var ydata yearData
	for rows.Next() {
		err = rows.Scan(&ydata.qty, &ydata.year)
		if err != nil {
			res.err = err
			return res
		}
		res.values = append(res.values, ydata)
	}

	res.err = rows.Err()
	return res
}

func (db *DB) addCategories(names ...string) error {
	stmt, err := db.Prepare("insert into categories(name) values(?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, name := range names {
		_, err = stmt.Exec(name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) getCategory(id int) (string, error) {
	var name string

	err := db.QueryRow("select name from categories where id = ?", id).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrInvalidCategory
		}
		return "", err
	}
	return name, nil
}

func (db *DB) getCategories() (map[int]string, error) {
	var (
		id   int
		name string

		res = make(map[int]string)
	)

	rows, err := db.Query("select id, name from categories")
	if err != nil {
		return res, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&id, &name)
		if err != nil {
			return res, err
		}
		res[id] = name
	}
	return res, rows.Err()
}

func (db *DB) addRecord(qty float64, category, year, week int) error {
	_, err := db.Exec("insert into records(qty, category, isoyear, isoweek) values(?, ?, ?, ?)",
		qty, category, year, week)
	return err
}

// Helpers
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

func open(p string) (*DB, error) {
	if !exists(p) {
		create(p)
	}

	db, err := sql.Open("sqlite3", p)
	if err != nil {
		return nil, err
	}
	err = db.Ping()

	return &DB{db}, err
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

func openFromContext(c *cli.Context) (*DB, error) {
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
