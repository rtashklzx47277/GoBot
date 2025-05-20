package sql

import (
	"GoBot/tools"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var userMap = map[string]map[string]string{
	"Youtube": {
		"UCrV1Hf5r8P148idjoSfrGEQ": "Sakuna",
		"UCLIpj4TmXviSTNE_U5WG_Ug": "Roa",
		"UC1opHUrw8rvnsadT-iGp7Cg": "Aqua",
		"UCXTpFs_3PqI41qX2d9tL2Rw": "Shion",
	},
	"Twitter": {
		"1512311952114028548": "Sakuna",
		"1850834672483459072": "Roa",
		"1024528894940987392": "Aqua",
		"1024533638879166464": "Shion",
		"1857716233757667335": "SakunaInfo",
		"1869731321217687552": "SakunaRadio",
	},
	"Fanbox": {
		"80355000": "Sakuna",
		"69014608": "Roa",
	},
	"Twitch": {
		"738746247": "Aqua",
		"773041510": "Shion",
	},
	"Twitcasting": {
		"1024528894940987392": "Aqua",
		"1024533638879166464": "Shion",
	},
	"Tiktok": {
		"minatoaqua_hololive":    "Aqua",
		"murasakishion_hololive": "Shion",
	},
}

type MySQL struct {
	Database *sql.DB
}

func ConnectToMySQL(username, password, host, port, dbName string) (*MySQL, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, dbName))
	if err != nil {
		return nil, fmt.Errorf("error connecting to MySQL!\n%v", err)
	}

	return &MySQL{db}, nil
}

func (mySQL *MySQL) Ping() error {
	return mySQL.Database.Ping()
}

func (mySQL *MySQL) Exec(query string, values ...any) {
	_, err := mySQL.Database.Exec(query, values...)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to exec query!\n%v", err))
	}
}

func (mySQL *MySQL) Insert(table string, data map[string]any) {
	var columns []string
	var values []any
	var placeholders []string

	for col, val := range data {
		columns = append(columns, col)
		values = append(values, val)
		placeholders = append(placeholders, "?")
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	_, err := mySQL.Database.Exec(query, values...)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to insert data!\n%v", err))
	}
}

func (mySQL *MySQL) Delete(table string, filter string, values ...any) {
	query := fmt.Sprintf("DELETE FROM %s %s", table, filter)
	_, err := mySQL.Database.Exec(query, values...)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to delete data!\n%v", err))
	}
}

func (mySQL *MySQL) Update(table, Id string, colAndVal ...any) {
	var columns []string
	var values []any

	for i := 0; i < len(colAndVal); i += 2 {
		columns = append(columns, fmt.Sprintf("%s = ?", colAndVal[i]))
		values = append(values, colAndVal[i+1])
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE Id = \"%s\"", table, strings.Join(columns, ", "), Id)
	_, err := mySQL.Database.Exec(query, values...)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to update data!\n%v", err))
	}
}

func (mySQL *MySQL) Find(table, filter string, values ...any) bool {
	var exist bool

	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s %s LIMIT 1)", table, filter)
	err := mySQL.Database.QueryRow(query, values...).Scan(&exist)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
	}

	return exist
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

func handleNullBool(d sql.NullBool) bool {
	if d.Valid {
		return d.Bool
	}

	return false
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
