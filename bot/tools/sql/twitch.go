package sql

import (
	"GoBot/tools/twitch"
	"database/sql"
	"fmt"
)

func (mySQL *MySQL) FindTwitchUser(userId string) twitch.User {
	var user twitch.User
	var loginId, title, description, channelTitle, color sql.NullString
	var followTime, slowTime sql.NullInt64

	query := "SELECT * FROM TwitchUser WHERE Id = ?"
	err := mySQL.db.QueryRow(query, userId).Scan(&user.Id, &loginId, &title, &description, &channelTitle, &color,
		&user.EmoteMode, &user.SubscriberMode, &user.UniqueMode,
		&user.FollowMode, &followTime, &user.SlowMode, &slowTime, &user.Live)
	if err != nil {
		fmt.Println(err)
	}

	user.LoginId = handleNullString(loginId)
	user.Title = handleNullString(title)
	user.Description = handleNullString(description)
	user.ChannelTitle = handleNullString(channelTitle)
	user.Color = handleNullString(color)
	user.FollowTime = handleNullInt(followTime)
	user.SlowTime = handleNullInt(slowTime)

	user.Icon = fmt.Sprintf("/bot/media/Twitch/%s/Icon/%s.jpg", user.Id, user.Id)
	user.Thumbnail = fmt.Sprintf("/bot/media/Twitch/%s/Thumbnail/%s.jpg", user.Id, user.Id)
	user.Url = fmt.Sprintf("https://www.twitch.tv/%s", user.LoginId)

	return user
}

func (mySQL *MySQL) FindTwitchSchedules(userId string) []twitch.Schedule {
	query := "SELECT ScheduledTime FROM TwitchSchedule WHERE UserId = ?"
	rows, err := mySQL.db.Query(query, userId)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var schedules []twitch.Schedule

	for rows.Next() {
		var schedule twitch.Schedule
		var scheduledTime sql.NullString

		err := rows.Scan(&scheduledTime)
		if err != nil {
			fmt.Println(err)
		}

		schedule.UserId = userId
		schedule.ScheduledTime = stringToTime(handleNullString(scheduledTime))

		schedules = append(schedules, schedule)
	}

	return schedules
}

func (mySQL *MySQL) FindTwitchBadges(userId string) []twitch.Badge {
	query := "SELECT Id, SetId, Title, Description FROM TwitchBadge WHERE UserId = ?"
	rows, err := mySQL.db.Query(query, userId)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var badges []twitch.Badge

	for rows.Next() {
		var badge twitch.Badge
		var setId, title, description sql.NullString

		err := rows.Scan(&badge.Id, &setId, &title, &description)
		if err != nil {
			fmt.Println(err)
		}

		badge.SetId = handleNullString(setId)
		badge.Title = handleNullString(title)
		badge.Description = handleNullString(description)

		badge.Image = fmt.Sprintf("/bot/media/Twitch/%s/Badge/%s.jpg", userId, badge.Id)

		badges = append(badges, badge)
	}

	return badges
}

func (mySQL *MySQL) FindTwitchStamps(userId string) []twitch.Stamp {
	query := "SELECT Id, Title, Tier, Type, Format FROM TwitchStamp WHERE UserId = ?"
	rows, err := mySQL.db.Query(query, userId)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var stamps []twitch.Stamp

	for rows.Next() {
		var stamp twitch.Stamp
		var title, tier, typeName, format sql.NullString

		err := rows.Scan(&stamp.Id, &title, &tier, &typeName, &format)
		if err != nil {
			fmt.Println(err)
		}

		stamp.Title = handleNullString(title)
		stamp.Tier = handleNullString(tier)
		stamp.Type = handleNullString(typeName)
		stamp.Format = handleNullString(format)

		stamp.Image = fmt.Sprintf("/bot/media/Twitch/%s/Stamp/%s.jpg", userId, stamp.Id)

		stamps = append(stamps, stamp)
	}

	return stamps
}
