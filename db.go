package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	ErrInvalidCategory = errors.New("category doesnt exist.")
	ErrNoName          = errors.New("tracker name required.")
)

type ErrInvalidDB struct {
	db string
}

func (err *ErrInvalidDB) Error() string {
	return fmt.Sprintf("tracker %s doesnt exist.", err.db)
}

type DB struct {
	*sql.DB
}

func (db *DB) query(q string, period Period) ([]timeData, error) {
	rows, err := db.Query(q)
	if err != nil {
		return []timeData{}, err
	}
	defer rows.Close()

	var (
		res  = make([]timeData, 0)
		data = timeData{period: period}

		datestr string
	)
	for rows.Next() {
		err = rows.Scan(&data.qty, &datestr)
		if err != nil {
			return res, err
		}

		data.date, err = time.Parse("2006-01-02", datestr)
		if err != nil {
			return res, err
		}
		res = append(res, data)
	}
	return res, rows.Err()
}

func (db *DB) queryDay(frequency int, categories []int) ([]timeData, error) {
	date := time.Now().AddDate(0, 0, -1*frequency)
	month, day := fmt.Sprintf("'%02d'", int(date.Month())), fmt.Sprintf("'%02d'", date.Day())
	qry := "select sum(records.qty) as quantity, records.date from records" +
		" where strftime('%Y', records.date) >= '" + itoa(date.Year()) + "'" +
		" and (strftime('%m', records.date) > " + month +
		" or (strftime('%m', records.date) = " + month +
		" and strftime('%d', records.date) >= " + day + ")) " + catCondition(categories) +
		"group by strftime('%Y-%m-%d', records.date)"

	return db.query(qry, DAY)
}

func (db *DB) queryWeek(frequency int, categories []int) ([]timeData, error) {
	date := time.Now().AddDate(0, 0, -7*frequency)
	year, week := date.ISOWeek()
	qry := "select sum(records.qty) as quantity, records.date from records " +
		"where ((strftime('%Y', date) = '" + itoa(year) + "'" +
		" and (strftime('%j', date(records.date, '-3 days', 'weekday 4')) - 1) / 7 + 1 >= " + itoa(week) + ") " +
		"or strftime('%Y', date) > '" + itoa(year) + "') " + catCondition(categories) +
		"group by strftime('%Y', date), (strftime('%j', date(records.date, '-3 days', 'weekday 4')) - 1) / 7 + 1"

	return db.query(qry, WEEK)
}

func (db *DB) queryMonth(frequency int, categories []int) ([]timeData, error) {
	date := time.Now().AddDate(0, -1*frequency, 0)
	year, month := date.Year(), int(date.Month())
	qry := "select sum(records.qty) as quantity, records.date from records " +
		"where ((strftime('%Y', records.date) = '" + itoa(year) + "' and strftime('%m', records.date) >= " + fmt.Sprintf("'%02d') ", month) +
		"or strftime('%Y', records.date) > '" + itoa(year) + "') " + catCondition(categories) +
		"group by strftime('%Y-%m', records.date)"

	return db.query(qry, MONTH)
}

func (db *DB) queryYear(frequency int, categories []int) ([]timeData, error) {
	date := time.Now().AddDate(-1*frequency, 0, 0)
	qry := "select sum(records.qty) as quantity, records.date from records " +
		"where strftime('%Y', records.date) >= '" + itoa(date.Year()) + "' " + catCondition(categories) +
		"group by strftime('%Y', records.date)"

	return db.query(qry, YEAR)
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

func (db *DB) addRecord(qty int64, category int) error {
	if _, err := db.getCategory(category); err != nil {
		return err
	}

	_, err := db.Exec("insert into records(qty, category) values(?, ?)", qty, category)
	return err
}

func catCondition(categories []int) string {
	var cond string

	if len(categories) > 0 {
		cond = fmt.Sprintf("and (records.category = %d", categories[0])
		for i := 1; i < len(categories); i++ {
			cond += fmt.Sprintf(" or records.category = %d", categories[i])
		}
		cond += ")"
	}
	return cond
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

func dblist() ([]string, error) {
	var names []string

	files, err := ioutil.ReadDir(TRACKER_DIR)
	if err != nil {
		return names, err
	}

	for _, f := range files {
		n := f.Name()
		if !f.IsDir() && strings.HasSuffix(n, ".db") {
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
	if _, err := os.Create(p); err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", p)
	if err != nil {
		return err
	}
	defer db.Close()

	exec := func(query string) {
		if err != nil {
			return
		}
		_, err = db.Exec(query)
	}

	exec("CREATE TABLE categories(id integer NOT NULL PRIMARY KEY AUTOINCREMENT, name text NOT NULL)")
	exec("CREATE TABLE records(id integer NOT NULL PRIMARY KEY AUTOINCREMENT, qty integer NOT NULL, " +
		"date integer NOT NULL DEFAULT CURRENT_DATE, category integer NOT NULL DEFAULT 1)")
	exec("INSERT INTO categories(name) VALUES('default')")
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
		return &ErrInvalidDB{dbname}
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
