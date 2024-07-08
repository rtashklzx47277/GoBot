package tiktok

import (
	"GoBot/tools"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func getData(path string) (*tools.Json, error) {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return &tools.Json{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &tools.Json{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return &tools.Json{}, fmt.Errorf("HTTP request failed with status code: %d\n%s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &tools.Json{}, err
	}

	match := tools.Regexp(string(body), `"user":(.*),"stats"`, 1)

	if len(match) == 0 {
		return &tools.Json{}, err
	}

	var jsonData tools.Json

	err = json.Unmarshal([]byte(match[0][1]), &jsonData.Data)
	if err != nil {
		return &tools.Json{}, err
	}

	return &jsonData, nil
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
