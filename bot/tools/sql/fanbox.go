package sql

import (
	"GoBot/tools/fanbox"
	"database/sql"
	"fmt"
)

func (mySQL *MySQL) FindFanboxUser(userId string) fanbox.User {
	var user fanbox.User
	var creatorId, name, description sql.NullString

	query := "SELECT * FROM FanboxUser WHERE Id = ?"
	err := mySQL.db.QueryRow(query, userId).Scan(&user.Id, &creatorId, &name, &description)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
	}

	user.CreatorId = handleNullString(creatorId)
	user.Name = handleNullString(name)
	user.Description = handleNullString(description)

	user.Icon = fmt.Sprintf("/bot/media/Fanbox/%s/Icon/%s.jpg", user.Id, user.Id)
	user.Banner = fmt.Sprintf("/bot/media/Fanbox/%s/Banner/%s.jpg", user.Id, user.Id)
	user.Url = fmt.Sprintf("https://www.fanbox.cc/@%s", user.CreatorId)

	return user
}

func (mySQL *MySQL) FindFanboxPlans(userId string) []fanbox.Plan {
	query := "SELECT Id, Title, Fee, Description FROM FanboxPlan WHERE UserId = ?"
	rows, err := mySQL.db.Query(query, userId)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
	}
	defer rows.Close()

	var plans []fanbox.Plan

	for rows.Next() {
		var plan fanbox.Plan
		var title, description sql.NullString
		var fee sql.NullInt64

		err := rows.Scan(&plan.Id, &title, &fee, &description)
		if err != nil {
			fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
		}

		plan.Title = handleNullString(title)
		plan.Fee = handleNullInt(fee)
		plan.Description = handleNullString(description)

		plan.Image = fmt.Sprintf("/bot/media/Fanbox/%s/Plan/%s.jpg", userId, plan.Id)

		plans = append(plans, plan)
	}

	return plans
}

func (mySQL *MySQL) FindFanboxPostIds(userId string) []string {
	query := "SELECT DISTINCT Id FROM FanboxPost WHERE UserId = ?"

	rows, err := mySQL.db.Query(query, userId)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
	}
	defer rows.Close()

	var result []string

	for rows.Next() {
		var value string

		err := rows.Scan(&value)
		if err != nil {
			fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
		}
		result = append(result, value)
	}

	return result
}
