package fanbox

import (
	"GoBot/tools"
	"fmt"
)

func getData(path string) (*tools.Json, error) {
	reader, err := tools.Get(path).AddHeader("origin", "https://www.fanbox.cc").Do()
	if err != nil {
		return &tools.Json{}, err
	}

	data, err := tools.ToJson(reader)
	if err != nil {
		return &tools.Json{}, err
	}

	return data, nil
}

func GetUser(userId string) (User, error) {
	url := fmt.Sprintf("https://api.fanbox.cc/creator.get?userId=%s", userId)
	data, err := getData(url)
	if err != nil {
		return User{}, err
	}

	item := data.Get("data").Get("body")
	user := User{
		Id:          item.Get("user").Get("userId").String(),
		CreatorId:   item.Get("creatorId").String(),
		Name:        item.Get("user").Get("name").String(),
		Description: item.Get("description").String(),
		Icon:        item.Get("user").Get("iconUrl").String(),
		Banner:      item.Get("coverImageUrl").String(),
	}

	user.Url = fmt.Sprintf("https://www.fanbox.cc/@%s", user.CreatorId)

	if user.Icon == "null" {
		user.Icon = tools.DefaultImage
	}

	if user.Banner == "null" {
		user.Banner = tools.DefaultImage
	}

	for _, link := range item.Get("profileLinks").Array() {
		user.Links = append(user.Links, link.(string))
	}

	for _, item := range item.Get("profileItems").JsonArray() {
		user.Items = append(user.Items, Item{
			Id:    item.Get("id").String(),
			Type:  item.Get("type").String(),
			Image: item.Get("imageUrl").String(),
		})
	}

	return user, nil
}

func GetPlan(userId string) ([]Plan, error) {
	url := fmt.Sprintf("https://api.fanbox.cc/plan.listCreator?userId=%s", userId)
	data, err := getData(url)
	if err != nil {
		return []Plan{}, err
	}

	var plans []Plan

	for _, item := range data.Get("body").JsonArray() {
		plan := Plan{
			Id:          item.Get("id").String(),
			UserId:      userId,
			Title:       item.Get("title").String(),
			Fee:         item.Get("fee").Int(),
			Description: item.Get("description").String(),
			Image:       item.Get("coverImageUrl").String(),
		}

		if plan.Image == "null" {
			plan.Image = tools.DefaultImage
		}

		plans = append(plans, plan)
	}

	return plans, nil
}

func GetPost(userId string) ([]Post, error) {
	url := fmt.Sprintf("https://api.fanbox.cc/post.listCreator?userId=%s&limit=100", userId)
	data, err := getData(url)
	if err != nil {
		return []Post{}, err
	}

	var posts []Post

	for _, item := range data.Get("body").Get("items").JsonArray() {
		post := Post{
			Id:            item.Get("id").String(),
			UserId:        userId,
			Title:         item.Get("title").String(),
			Fee:           item.Get("feeRequired").Int(),
			Image:         item.Get("cover").String(),
			PublishedTime: item.Get("publishedDatetime").Time(),
			UpdatedTime:   item.Get("updatedDatetime").Time(),
		}

		post.Url = fmt.Sprintf("https://www.fanbox.cc/@%s/posts/%s", item.Get("creatorId").String(), post.Id)

		if post.Image == "null" {
			post.Image = tools.DefaultImage
		}

		posts = append(posts, post)
	}

	return posts, nil
}
