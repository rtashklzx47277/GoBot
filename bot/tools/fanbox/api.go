package fanbox

import (
	"GoBot/tools"
	"fmt"
)

func getData(path string) (*tools.Json, error) {
	reader, err := tools.Get(path).
		AddHeader("Origin", "https://www.fanbox.cc").Do()
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
	url := fmt.Sprintf("https://api.fanbox.cc/creator.get?creatorId=%s", userId)
	data, err := getData(url)
	if err != nil {
		return User{}, err
	}

	item := data.Get("body")
	user := User{
		Id:          item.Get("user").Get("userId").String(),
		CreatorId:   item.Get("creatorId").String(),
		Name:        item.Get("user").Get("name").String(),
		Description: item.Get("description").String(),
		Icon:        item.Get("user").Get("iconUrl").String(),
		Category:    item.Get("category").String(),
		Banner:      item.Get("coverImageUrl").String(),
	}

	user.Url = fmt.Sprintf("https://www.pixiv.net/fanbox/creator/%s", user.Id)

	if user.Icon == "null" {
		user.Icon = ""
	}

	if user.Banner == "null" {
		user.Banner = ""
	}

	for _, link := range item.Get("profileLinks").Array() {
		user.Links = append(user.Links, link.(string))
	}

	for _, thing := range item.Get("profileItems").JsonArray() {
		item := Item{
			Id:     thing.Get("id").String(),
			UserId: user.Id,
			Type:   thing.Get("type").String(),
		}

		if item.Type == "image" {
			item.Media = thing.Get("imageUrl").String()
		} else if item.Type == "video" {
			videoId := thing.Get("videoId").String()

			switch thing.Get("serviceProvider").String() {
			case "youtube":
				item.Media = fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoId)
			case "vimeo":
				item.Media = fmt.Sprintf("https://vimeo.com/%s", videoId)
			case "soundcloud":
				item.Media = fmt.Sprintf("https://soundcloud.com/%s", videoId)
			}
		}

		user.Items = append(user.Items, item)
	}

	return user, nil
}

func GetPlan(userId string) ([]Plan, error) {
	url := fmt.Sprintf("https://api.fanbox.cc/plan.listCreator?creatorId=%s", userId)
	data, err := getData(url)
	if err != nil {
		return []Plan{}, err
	}

	var plans []Plan

	for _, item := range data.Get("body").JsonArray() {
		plan := Plan{
			Id:          item.Get("id").String(),
			UserId:      item.Get("user").Get("userId").String(),
			Title:       item.Get("title").String(),
			Fee:         item.Get("fee").Int(),
			Description: item.Get("description").String(),
			Image:       item.Get("coverImageUrl").String(),
		}

		plans = append(plans, plan)
	}

	return plans, nil
}

func GetPost(userId string) ([]Post, error) {
	url := fmt.Sprintf("https://api.fanbox.cc/post.listCreator?creatorId=%s&limit=300", userId)
	data, err := getData(url)
	if err != nil {
		return []Post{}, err
	}

	var posts []Post

	for _, item := range data.Get("body").JsonArray() {
		post := Post{
			Id:            item.Get("id").String(),
			UserId:        item.Get("user").Get("userId").String(),
			Title:         item.Get("title").String(),
			Fee:           item.Get("feeRequired").Int(),
			Image:         item.Get("cover").Get("url").String(),
			PublishedTime: item.Get("publishedDatetime").Time(),
			UpdatedTime:   item.Get("updatedDatetime").Time(),
		}

		post.Url = fmt.Sprintf("https://www.fanbox.cc/@%s/posts/%s", item.Get("creatorId").String(), post.Id)

		posts = append(posts, post)
	}

	return posts, nil
}
