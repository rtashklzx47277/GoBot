package sql

import (
	"GoBot/tools/twitcasting"
	"database/sql"
	"fmt"
)

func (mySQL *MySQL) FindTwitcastingUser(userId string) twitcasting.User {
	var user twitcasting.User
	var screenId, title, description sql.NullString

	query := "SELECT * FROM TwitcastingUser WHERE Id = ?"
	err := mySQL.db.QueryRow(query, userId).Scan(&user.Id, &screenId, &title, &description, &user.Live)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
	}

	user.ScreenId = handleNullString(screenId)
	user.Title = handleNullString(title)
	user.Description = handleNullString(description)

	user.Icon = fmt.Sprintf("/bot/media/%s/Twitcasting/Icon.jpg", userMap["Twitcasting"][userId])
	user.Url = fmt.Sprintf("https://twitcasting.tv/%s", user.ScreenId)

	return user
}
