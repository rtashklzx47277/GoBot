package tools

import (
	"encoding/json"
	"os"

	"github.com/bwmarrin/discordgo"
)

var (
	UserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36 Edg/131.0.0.0"
	ClientVersion = "2.20240620.05.00"

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "collab",
			Description: "New Collab Stream",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "collab-with",
					Description: "Which User Is Collabing",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Sakuna",
							Value: "Sakuna",
						},
						{
							Name:  "Roa",
							Value: "Roa",
						},
						{
							Name:  "Both",
							Value: "Both",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "video-id",
					Description: "Youtube Video Id",
					Required:    true,
				},
			},
		},
		{
			Name:        "follow",
			Description: "Follow User",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "follow-by",
					Description: "Which User Is Following",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Sakuna",
							Value: "Sakuna",
						},
						{
							Name:  "Roa",
							Value: "Roa",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "usernames",
					Description: "Twitter User Screen Names (separate multiple usernames with commas)",
					Required:    true,
				},
			},
		},
		{
			Name:        "unfollow",
			Description: "Unfollow User",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "unfollow-by",
					Description: "Which User Is Unfollowing",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "Sakuna",
							Value: "Sakuna",
						},
						{
							Name:  "Roa",
							Value: "Roa",
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "usernames",
					Description: "Twitter User Screen Names (separate multiple usernames with commas)",
					Required:    true,
				},
			},
		},
		{
			Name:        "update-youtube-channels-data",
			Description: "Update Youtube Channels Data",
		},
		{
			Name:        "update-twitter-users-data",
			Description: "Update Twitter Users Data",
		},
		{
			Name:        "get-twitch-access-token",
			Description: "Return Twitch Access Token",
		},
	}
)

type UserData map[string]Data

type Data struct {
	Youtube struct {
		Id               string `json:"Id"`
		DiscordChannelId string `json:"DiscordChannelId"`
	} `json:"Youtube"`
	Twitter struct {
		Id               string `json:"Id"`
		Username         string `json:"Username"`
		DiscordChannelId string `json:"DiscordChannelId"`
	} `json:"Twitter"`
	Twitch struct {
		Id               string `json:"Id"`
		DiscordChannelId string `json:"DiscordChannelId"`
	} `json:"Twitch"`
	Twitcasting struct {
		Id               string `json:"Id"`
		DiscordChannelId string `json:"DiscordChannelId"`
	} `json:"Twitcasting"`
	Tiktok struct {
		Id               string `json:"Id"`
		DiscordChannelId string `json:"DiscordChannelId"`
	} `json:"Tiktok"`
	Fanbox struct {
		Id               string `json:"Id"`
		CreatorId        string `json:"CreatorId"`
		DiscordChannelId string `json:"DiscordChannelId"`
	} `json:"Fanbox"`
}

func GetUserData() (map[string]Data, error) {
	data, err := os.ReadFile("/bot/userData.json")
	if err != nil {
		return map[string]Data{}, err
	}

	var userData map[string]Data

	err = json.Unmarshal(data, &userData)
	if err != nil {
		return map[string]Data{}, err
	}

	return userData, nil
}

func GetUserMaps(userData map[string]Data, names ...string) ([]string, map[string]map[string]string) {
	var usernames []string
	userMap := make(map[string]map[string]string)

	for _, name := range names {
		usernames = append(usernames, userData[name].Twitter.Username)
		userMap[userData[name].Twitter.Id] = map[string]string{"Name": name, "DiscordChannelId": userData[name].Twitter.DiscordChannelId}
	}

	return usernames, userMap
}

func GetUsername(userData map[string]Data, userId string) string {
	for _, data := range userData {
		if data.Twitter.Id == userId {
			return data.Twitter.Username
		}
	}

	return ""
}
