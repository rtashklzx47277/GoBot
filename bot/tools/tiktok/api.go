package tiktok

import (
	"GoBot/tools"
	"fmt"
)

func getData(path string) (*tools.Json, error) {
	reader, err := tools.Get(path).Do()
	if err != nil {
		return &tools.Json{}, err
	}

	data, err := tools.ToString(reader)

	match, ok := tools.Regexp(data, `"user":(.*),"stats"`)
	if !ok {
		return &tools.Json{}, fmt.Errorf("failed to get user data!\n%w", err)
	}

	jsonData, err := tools.StringToJson(match)
	if err != nil {
		return &tools.Json{}, err
	}

	return jsonData, nil
}

func GetUser(userId string) (User, error) {
	url := fmt.Sprintf("https://www.tiktok.com/@%s", userId)
	data, err := getData(url)
	if err != nil {
		return User{}, err
	}

	user := User{
		Id:          data.Get("id").String(),
		ShortId:     data.Get("shortId").String(),
		UniqueId:    data.Get("uniqueId").String(),
		Title:       data.Get("nickname").String(),
		Description: data.Get("signature").String(),
		Icon:        data.Get("avatarLarger").String(),
	}

	user.Url = fmt.Sprintf("https://www.tiktok.com/@%s", user.UniqueId)

	return user, nil
}
