package twitch

import (
	"GoBot/tools"
	"bytes"
	"fmt"
	"net/url"
	"os"
)

var (
	clientId     = os.Getenv("TIWTCH_CLIENT_ID")
	clientSecret = os.Getenv("TIWTCH_CLIENT_SECRET")
	accessToken  = os.Getenv("TIWTCH_ACCESS_TOKEN")
)

func getData(path string) (*tools.Json, error) {
	reader, err := tools.Get(path).
		AddHeader("Client-Id", clientId).
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
	url := fmt.Sprintf("https://api.twitch.tv/helix/users?id=%s", userId)
	data, err := getData(url)
	if err != nil {
		return User{}, fmt.Errorf("failed to get user data!\n%w", err)
	}

	item := data.Get("data").Index(0)
	user := User{
		Id:          item.Get("id").String(),
		LoginId:     item.Get("login").String(),
		Title:       item.Get("display_name").String(),
		Description: item.Get("description").String(),
		Icon:        item.Get("profile_image_url").String(),
		Thumbnail:   item.Get("offline_image_url").String(),
	}

	user.Url = fmt.Sprintf("https://www.twitch.tv/%s", user.LoginId)

	if user.Thumbnail == "None" {
		user.Thumbnail = tools.DefaultImage
	}

	url = fmt.Sprintf("https://api.twitch.tv/helix/channels?broadcaster_id=%s", userId)
	data, err = getData(url)
	if err != nil {
		return User{}, fmt.Errorf("failed to get channel data!\n%w", err)
	}

	user.ChannelTitle = data.Get("data").Index(0).Get("title").String()

	url = fmt.Sprintf("https://api.twitch.tv/helix/chat/color?user_id=%s", userId)
	data, err = getData(url)
	if err != nil {
		return User{}, fmt.Errorf("failed to get chat color!\n%w", err)
	}

	user.Color = data.Get("data").Index(0).Get("color").String()

	url = fmt.Sprintf("https://api.twitch.tv/helix/chat/settings?broadcaster_id=%s", userId)
	data, err = getData(url)
	if err != nil {
		return User{}, fmt.Errorf("failed to get chat setting!\n%w", err)
	}

	item = data.Get("data").Index(0)

	user.EmoteMode = item.Get("emote_mode").Bool()
	user.SubscriberMode = item.Get("subscriber_mode").Bool()
	user.UniqueMode = item.Get("unique_chat_mode").Bool()
	user.FollowMode = item.Get("follower_mode").Bool()
	user.FollowTime = item.Get("follower_mode_duration").Int()
	user.SlowMode = item.Get("slow_mode").Bool()
	user.SlowTime = item.Get("slow_mode_wait_time").Int()

	return user, nil
}

func GetLive(userId string) (bool, string, error) {
	url := fmt.Sprintf("https://api.twitch.tv/helix/streams?user_id=%s", userId)
	data, err := getData(url)
	if err != nil {
		return false, "", err
	}

	if len(data.Get("data").JsonArray()) == 0 {
		return false, "", nil
	}

	return true, data.Get("data").Index(0).Get("title").String(), nil
}

func GetVideos(userId string) ([]Video, error) {
	url := fmt.Sprintf("https://api.twitch.tv/helix/videos?user_id=%s", userId)
	data, err := getData(url)
	if err != nil {
		return []Video{}, err
	}

	var videos []Video

	for _, item := range data.Get("data").JsonArray() {
		video := Video{
			Id:          item.Get("id").String(),
			StreamId:    item.Get("stream_id").String(),
			Title:       item.Get("title").String(),
			Description: item.Get("description").String(),
			Type:        item.Get("type").String(),
			Length:      item.Get("duration").Duration(),
			Created:     item.Get("created_at").Time(),
			Published:   item.Get("published_at").Time(),
		}

		videos = append(videos, video)
	}

	return videos, nil
}

func GetBadges(userId string) ([]Badge, error) {
	url := fmt.Sprintf("https://api.twitch.tv/helix/chat/badges?broadcaster_id=%s", userId)
	data, err := getData(url)
	if err != nil {
		return []Badge{}, err
	}

	var badges []Badge

	for _, set := range data.Get("data").JsonArray() {
		setId := set.Get("set_id").String()

		for _, item := range set.Get("versions").JsonArray() {
			badge := Badge{
				Id:          item.Get("id").String(),
				UserId:      userId,
				SetId:       setId,
				Title:       item.Get("title").String(),
				Description: item.Get("description").String(),
				Image:       item.Get("image_url_4x").String(),
			}

			badges = append(badges, badge)
		}
	}

	return badges, nil
}

func GetStamps(userId string) ([]Stamp, error) {
	url := fmt.Sprintf("https://api.twitch.tv/helix/chat/emotes?broadcaster_id=%s", userId)
	data, err := getData(url)
	if err != nil {
		return []Stamp{}, err
	}

	var stamps []Stamp

	for _, item := range data.Get("data").JsonArray() {
		stamp := Stamp{
			Id:     item.Get("id").String(),
			UserId: userId,
			Title:  item.Get("name").String(),
			Tier:   item.Get("tier").String(),
			Type:   item.Get("emote_type").String(),
			Format: item.Get("description").String(),
			Image:  item.Get("images").Get("url_4x").String(),
		}

		if len(item.Get("format").Array()) == 2 {
			stamp.Format = "both"
		} else {
			stamp.Format = item.Get("format").Array()[0].(string)
		}

		stamps = append(stamps, stamp)
	}

	return stamps, nil
}

func GetSchedule(userId string) ([]Schedule, error) {
	url := fmt.Sprintf("https://api.twitch.tv/helix/schedule?broadcaster_id=%s", userId)
	data, err := getData(url)
	if err != nil {
		return []Schedule{}, err
	}

	var schedules []Schedule

	for _, item := range data.Get("data").JsonArray() {
		stamp := Schedule{
			UserId:        userId,
			ScheduledTime: item.Get("start_time").Time(),
		}

		schedules = append(schedules, stamp)
	}

	return schedules, nil
}

func GetAccessToken() (string, error) {
	body := url.Values{}
	body.Set("client_id", clientId)
	body.Set("client_secret", clientSecret)
	body.Set("grant_type", "client_credentials")

	reader, err := tools.Post("https://id.twitch.tv/oauth2/token", bytes.NewBufferString(body.Encode())).
		AddHeader("Content-Type", "application/x-www-form-urlencoded").Do()
	if err != nil {
		return "", err
	}

	data, err := tools.ToJson(reader)
	if err != nil {
		return "", err
	}

	return data.Get("access_token").String(), nil
}
