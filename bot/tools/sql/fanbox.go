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
