package twitcasting

import (
	"GoBot/tools"
	"fmt"
	"os"
)

var accessToken = os.Getenv("TWITCASTING_ACCESS_TOKEN")

func getData(path string) (*tools.Json, error) {
	reader, err := tools.Get(path).
		AddHeader("Accept", "application/json").
		AddHeader("X-Api-Version", "2.0").
		AddHeader("Authorization", fmt.Sprintf("Bearer %s", accessToken)).Do()
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
	url := fmt.Sprintf("https://apiv2.twitcasting.tv/users/%s", userId)
	data, err := getData(url)
	if err != nil {
		return User{}, err
	}

	item := data.Get("user")
	user := User{
		Id:          item.Get("id").String(),
		ScreenId:    item.Get("screen_id").String(),
		Title:       item.Get("name").String(),
		Description: item.Get("profile").String(),
		Icon:        item.Get("image").String(),
		Live:        item.Get("is_live").Bool(),
	}

	user.Url = fmt.Sprintf("https://twitcasting.tv/%s", user.ScreenId)

	return user, nil
}

func GetStream(userId string) (bool, string, error) {
	user, err := GetUser(userId)
	if err != nil {
		return false, "", err
	}

	if !user.Live {
		return false, "", nil
	}

	url := fmt.Sprintf("https://apiv2.twitcasting.tv/users/%s", userId)
	data, err := getData(url)
	if err != nil {
		return false, "", err
	}

	streamId := data.Get("user").Get("last_movie_id")
	url = fmt.Sprintf("https://apiv2.twitcasting.tv/movies/%s", streamId)
	data, err = getData(url)
	if err != nil {
		return false, "", err
	}

	title := fmt.Sprintf("%s %s", data.Get("movie").Get("title").String(), data.Get("movie").Get("subtitle").String())

	return true, title, nil
}
