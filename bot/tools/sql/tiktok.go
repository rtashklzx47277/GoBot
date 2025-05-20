package sql

import (
	"GoBot/tools/tiktok"
	"database/sql"
	"fmt"
)

func (mySQL *MySQL) FindTiktokUser(userId string) tiktok.User {
	var user tiktok.User
	var shortId, uniqueId, title, description sql.NullString
	var followCount sql.NullInt64

	query := "SELECT Id, ShortId, UniqueId, Title, Description, FollowCount FROM TiktokUser WHERE Id = ?"
	err := mySQL.Database.QueryRow(query, userId).Scan(&user.Id, &shortId, &uniqueId, &title, &description, &followCount)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
	}

	user.ShortId = handleNullString(shortId)
	user.UniqueId = handleNullString(uniqueId)
	user.Title = handleNullString(title)
	user.Description = handleNullString(description)
	user.FollowCount = handleNullInt(followCount)

	user.Icon = fmt.Sprintf("/bot/media/%s/Tiktok/Icon.jpg", userMap["Tiktok"][userId])
	user.Url = fmt.Sprintf("https://www.tiktok.com/@%s", user.UniqueId)

	return user
}
