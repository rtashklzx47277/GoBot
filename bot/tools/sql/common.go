package sql

import (
	"GoBot/tools"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQL struct {
	db *sql.DB
}

func ConnectToMySQL(username, password, host, port, dbName string) (*MySQL, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbName))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &MySQL{db}, nil
}

func (mySQL *MySQL) Insert(table string, data map[string]any) {
	var values []any
	query := "INSERT INTO " + table + " ("

	for key, val := range data {
		query += key + ", "
		values = append(values, val)
	}

	query = query[:len(query)-2] + ") VALUES ("

	for i := 0; i < len(values); i++ {
		query += "?, "
	}

	query = query[:len(query)-2] + ")"

	_, err := mySQL.db.Exec(query, values...)
	if err != nil {
		fmt.Println(err)
	}
}

func (mySQL *MySQL) Delete(table string, filter string, values ...any) {
	query := fmt.Sprintf("DELETE FROM %s %s", table, filter)
	_, err := mySQL.db.Exec(query, values...)
	if err != nil {
		fmt.Println(err)
	}
}

func (mySQL *MySQL) Update(table, Id string, colAndVal ...any) {
	var colArray []string
	var valArray []any

	for i := 0; i < len(colAndVal); i += 2 {
		colArray = append(colArray, fmt.Sprintf("%s = ?", colAndVal[i]))
		valArray = append(valArray, colAndVal[i+1])
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE Id = \"%s\"", table, strings.Join(colArray, ", "), Id)
	_, err := mySQL.db.Exec(query, valArray...)
	if err != nil {
		fmt.Println(err)
	}
}

func (mySQL *MySQL) Find(table, filter string, values ...any) bool {
	var count int

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", table, filter)
	err := mySQL.db.QueryRow(query, values...).Scan(&count)
	if err != nil {
		fmt.Println(err)
	}

	return count != 0
}

func handleNullString(d sql.NullString) string {
	if d.Valid {
		return d.String
	}

	return "None"
}

func handleNullInt(d sql.NullInt64) int {
	if d.Valid {
		return int(d.Int64)
	}

	return 0
}

func stringToDuration(s string) tools.Duration {
	ds, err := time.ParseDuration(fmt.Sprintf("%ss", strings.Replace(strings.Replace(s, ":", "h", 1), ":", "m", 1)))
	if err != nil {
		return tools.Duration(0)
	}

	return tools.Duration(ds)
}

func stringToTime(s string) tools.Time {
	t, err := time.Parse("2006-01-02 15:04:05", string(s))
	if err != nil {
		return tools.Time{}
	}

	return tools.Time(t)
}
