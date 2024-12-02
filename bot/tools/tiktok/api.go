package tiktok

import (
	"GoBot/tools"
	"fmt"
)

var ErrorNoUserData = fmt.Errorf("failed to get user data")

func getData(path string) (string, error) {
	reader, err := tools.Get(path).Do()
	if err != nil {
		return "", err
	}

	data, err := tools.ToString(reader)
	if err != nil {
		return "", err
	}

	return data, nil
}

func GetUser(userId string) (User, error) {
	url := fmt.Sprintf("https://www.tiktok.com/@%s", userId)
	data, err := getData(url)
	if err != nil {
		return User{}, err
	}

	match, ok := tools.Regexp(data, `"user":(.*),"stats"`)
	if !ok {
		return User{}, ErrorNoUserData
	}

	jsonData, err := tools.StringToJson(match)
	if err != nil {
		return User{}, err
	}

	user := User{
		Id:          jsonData.Get("id").String(),
		ShortId:     jsonData.Get("shortId").String(),
		UniqueId:    jsonData.Get("uniqueId").String(),
		Title:       jsonData.Get("nickname").String(),
		Description: jsonData.Get("signature").String(),
		Icon:        jsonData.Get("avatarLarger").String(),
	}

	user.Url = fmt.Sprintf("https://www.tiktok.com/@%s", user.UniqueId)

	match, ok = tools.Regexp(data, `"stats":(.*),"itemList"`)
	if !ok {
		return User{}, fmt.Errorf("failed to get user data!\n%w", err)
	}

	jsonData, err = tools.StringToJson(match)
	if err != nil {
		return User{}, err
	}

	user.FollowCount = jsonData.Get("followingCount").Int()

	return user, nil
}
