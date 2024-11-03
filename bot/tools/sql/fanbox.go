package sql

import (
	"GoBot/tools/fanbox"
	"database/sql"
	"fmt"
)

func (mySQL *MySQL) FindFanboxUser(userId string) fanbox.User {
	var user fanbox.User
	var creatorId, name, description, category sql.NullString

	query := "SELECT Id, CreatorId, Name, Description, Category FROM FanboxUser WHERE Id = ?"
	err := mySQL.db.QueryRow(query, userId).Scan(&user.Id, &creatorId, &name, &description, &category)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
	}

	user.CreatorId = handleNullString(creatorId)
	user.Name = handleNullString(name)
	user.Description = handleNullString(description)
	user.Category = handleNullString(category)

	user.Icon = fmt.Sprintf("/bot/media/Fanbox/%s/Icon/%s.jpg", user.Id, user.Id)
	user.Banner = fmt.Sprintf("/bot/media/Fanbox/%s/Banner/%s.jpg", user.Id, user.Id)
	user.Url = fmt.Sprintf("https://www.fanbox.cc/@%s", user.CreatorId)

	query = "SELECT Link FROM FanboxLink WHERE UserId = ?"
	rows, err := mySQL.db.Query(query, userId)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
	}
	defer rows.Close()

	for rows.Next() {
		var link string

		err := rows.Scan(&link)
		if err != nil {
			fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
		}

		user.Links = append(user.Links, link)
	}

	query = "SELECT Id, Type, Media FROM FanboxItem WHERE UserId = ?"
	rows, err = mySQL.db.Query(query, userId)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
	}
	defer rows.Close()

	for rows.Next() {
		var item fanbox.Item
		var class, media sql.NullString

		err := rows.Scan(&item.Id, &class, &media)
		if err != nil {
			fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
		}

		item.Type = handleNullString(class)
		item.Media = handleNullString(media)

		user.Items = append(user.Items, item)
	}

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

func (mySQL *MySQL) FindFanboxPosts(userId string) []fanbox.Post {
	query := "SELECT Id, UserId, Title, Fee, UpdatedTime FROM FanboxPost WHERE UserId = ?"
	rows, err := mySQL.db.Query(query, userId)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
	}
	defer rows.Close()

	var posts []fanbox.Post

	for rows.Next() {
		var post fanbox.Post
		var title, updatedTime sql.NullString
		var fee sql.NullInt64

		err := rows.Scan(&post.Id, &post.UserId, &title, &fee, &updatedTime)
		if err != nil {
			fmt.Println(fmt.Errorf("failed to find data!\n%v", err))
		}

		post.Title = handleNullString(title)
		post.Fee = handleNullInt(fee)
		post.UpdatedTime = stringToTime(handleNullString(updatedTime))

		post.Image = fmt.Sprintf("/bot/media/Fanbox/%s/Post/%s.jpg", userId, post.Id)

		posts = append(posts, post)
	}

	return posts
}
