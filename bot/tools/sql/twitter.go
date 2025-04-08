package sql

import (
	"GoBot/tools/twitter"
	"database/sql"
	"fmt"
)

func (mySQL *MySQL) FindTwitterUser(userId string) twitter.User {
	var user twitter.User
	var username, name, description, location, pinned sql.NullString

	query := "SELECT * FROM TwitterUser WHERE Id = ?"
	err := mySQL.db.QueryRow(query, userId).Scan(&user.Id, &username, &name, &description, &location, &pinned,
		&user.Protected, &user.Verified, &user.FollowersCount, &user.FollowingCount, &user.LikeCount)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
	}

	user.Username = handleNullString(username)
	user.Name = handleNullString(name)
	user.Description = handleNullString(description)
	user.Location = handleNullString(location)
	user.Pinned = handleNullString(pinned)

	user.Icon = fmt.Sprintf("/bot/media/Twitter/%s/Icon/%s.jpg", user.Id, user.Id)
	user.Banner = fmt.Sprintf("/bot/media/Twitter/%s/Banner/%s.jpg", user.Id, user.Id)
	user.Url = fmt.Sprintf("https://x.com/%s", user.Username)

	return user
}
