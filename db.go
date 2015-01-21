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

func (db *DB) queryWeek(occurence, category int) ([]dataFormatter, error) {
	var (
		year, week = time.Now().ISOWeek()
		condition  = catCondition(category)
		res        = make([]dataFormatter, 0)

		query string
	)

	if occurence == 0 {
		// query concerns current week
		query = fmt.Sprintf("select quantity, isoyear, isoweek from ("+
			"select sum(qty) as quantity, isoyear, isoweek from records "+
			"where isoyear = %d and isoweek = %d %sgroup by isoweek"+
			") where quantity is not null", year, week, condition)
	} else {
		year, week = computeLimitWeek(year, week, occurence)

		query = fmt.Sprintf("select quantity, isoyear, isoweek from ("+
			"select sum(qty) as quantity, isoyear, isoweek from records "+
			"where ((isoyear >= %d and isoweek >= %d) "+
			"or isoyear > %d) %sgroup by isoweek) where quantity is not null", year, week, year, condition)
	}

	rows, err := db.Query(query)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	var wdata weekData
	for rows.Next() {
		err = rows.Scan(&wdata.qty, &wdata.year, &wdata.week)
		if err != nil {
			return res, err
		}
		res = append(res, wdata)
	}

	return res, rows.Err()
}

func (db *DB) queryMonth(occurence, category int) ([]dataFormatter, error) {
	var (
		date        = time.Now()
		year, month = date.Year(), date.Month()
		condition   = catCondition(category)
		res         = make([]dataFormatter, 0)

		query string
	)

	if occurence == 0 {
		query = "select quantity, isoyear, month from (" +
			"select sum(qty) as quantity, isoyear, strftime('%m', date) as month " +
			"from records where strftime('%m', date) = " + fmt.Sprintf("'%s' ", monthVal(int(month))) +
			"and strftime('%Y', date) = " + fmt.Sprintf("'%s' ", strconv.Itoa(year)) + condition +
			") where quantity is not null"
	} else {
		y, m := computeLimitMonth(year, int(month), occurence)
		ystr, mstr := fmt.Sprintf("'%s' ", strconv.Itoa(y)), fmt.Sprintf("'%s' ", monthVal(int(m)))

		query = "select quantity, isoyear, month from (" +
			"select sum(qty) as quantity, isoyear, strftime('%m', date) as month " +
			"from records where ((strftime('%Y', date) >= " + ystr +
			"and strftime('%m', date) >= " + mstr + ")" +
			"or strftime('%Y', date) > " + ystr + ")" + condition +
			"group by strftime('%Y-%m', date)" +
			") where quantity is not null"
	}

	rows, err := db.Query(query)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	var mdata monthData
	for rows.Next() {
		err = rows.Scan(&mdata.qty, &mdata.year, &mdata.month)
		if err != nil {
			return res, err
		}
		res = append(res, mdata)
	}

	return res, rows.Err()
}

func (db *DB) queryYear(occurence, category int) ([]dataFormatter, error) {
	var (
		year      = time.Now().Year()
		condition = catCondition(category)
		res       = make([]dataFormatter, 0)

		query string
	)

	if occurence == 0 {
		query = fmt.Sprintf("select quantity, isoyear from ("+
			"select sum(qty) as quantity, isoyear from records "+
			"where isoyear = %d %s) where quantity is not null", year, condition)
	} else {
		year = year - occurence

		query = fmt.Sprintf("select quantity, isoyear from ("+
			"select sum(qty) as quantity, isoyear from records "+
			"where isoyear >= %d %sgroup by isoyear) where quantity is not null", year, condition)
	}

	rows, err := db.Query(query)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	var ydata yearData
	for rows.Next() {
		err = rows.Scan(&ydata.qty, &ydata.year)
		if err != nil {
			return res, err
		}
		res = append(res, ydata)
	}

	return res, rows.Err()
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

func catCondition(category int) (cond string) {
	if category > 0 {
		cond = fmt.Sprintf("and category = %d ", category)
	}
	return
}

func monthVal(m int) string {
	str := strconv.Itoa(m)
	if m < 10 {
		str = "0" + str
	}
	return str
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

func withDBContext(dbname string, fn func(*DB) error) error {
	if dbname == "" {
		return ErrNoName
	}

	p := dbpath(dbname)
	if !exists(p) {
		return ErrInvalidDB
	}

	db, err := open(p)
	if err != nil {
		return err
	}
	defer db.Close()

	return fn(db)
}

func dbpath(name string) string {
	return path.Join(TRACKER_DIR, name+".db")
}
