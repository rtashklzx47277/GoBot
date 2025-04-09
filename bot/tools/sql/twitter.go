package sql

import (
	"GoBot/tools/twitter"
	"database/sql"
	"fmt"
)

func (mySQL *MySQL) FindTwitterUser(userId string) twitter.User {
	var user twitter.User
	var username, name, description, location, link, latest, pinned sql.NullString

	query := "SELECT * FROM TwitterUser WHERE Id = ?"
	err := mySQL.db.QueryRow(query, userId).Scan(&user.Id, &username, &name, &description, &location, &link, &latest, &pinned,
		&user.Protected, &user.Verified, &user.FollowersCount, &user.FollowingCount, &user.LikeCount)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
	}

	user.Username = handleNullString(username)
	user.Name = handleNullString(name)
	user.Description = handleNullString(description)
	user.Location = handleNullString(location)
	user.Link = handleNullString(link)
	user.Latest = handleNullString(latest)
	user.Pinned = handleNullString(pinned)

	user.Icon = fmt.Sprintf("/bot/media/%s/Twitter/Icon.jpg", userMap["Twitter"][user.Id])
	user.Banner = fmt.Sprintf("/bot/media/%s/Twitter/Banner.jpg", userMap["Twitter"][user.Id])
	user.Url = fmt.Sprintf("https://x.com/%s", user.Username)

	return user
}
