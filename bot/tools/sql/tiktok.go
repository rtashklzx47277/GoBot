package sql

import (
	"GoBot/tools/tiktok"
	"database/sql"
	"fmt"
)

func (mySQL *MySQL) FindTiktokUser(userId string) tiktok.User {
	var user tiktok.User
	var shortId, uniqueId, title, description sql.NullString

	query := "SELECT * FROM TiktokUser WHERE Id = ?"
	err := mySQL.db.QueryRow(query, userId).Scan(&user.Id, &shortId, &uniqueId, &title, &description)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
	}

	user.ShortId = handleNullString(shortId)
	user.UniqueId = handleNullString(uniqueId)
	user.Title = handleNullString(title)
	user.Description = handleNullString(description)

	user.Icon = fmt.Sprintf("/bot/media/Tiktok/%s/Icon/%s.jpg", user.Id, user.Id)
	user.Url = fmt.Sprintf("https://www.tiktok.com/@%s", user.UniqueId)

	return user
}
