package sql

import "fmt"

func (mySQL *MySQL) FindArticleId() (string, string) {
	var HPId, radioId string

	query := "SELECT Id FROM Article WHERE `From` = ? ORDER BY Id DESC LIMIT 1"
	err := mySQL.db.QueryRow(query, "HP").Scan(&HPId)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
	}

	query = "SELECT Id FROM Article WHERE `From` = ? ORDER BY Date DESC LIMIT 1"
	err = mySQL.db.QueryRow(query, "Radio").Scan(&radioId)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
	}

	return HPId, radioId
}
