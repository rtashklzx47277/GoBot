package main

import (
	"GoBot/tools"
	"GoBot/tools/discord"
	"GoBot/tools/sql"
	"GoBot/tools/twitch"
	"GoBot/tools/youtube"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	db            *sql.MySQL
	s             *discordgo.Session
	collabIds     = []string{}
	messageIdList = []string{}
	testChannelId = os.Getenv("DISCORD_TEST_CHANNEL_ID")
)

var userDataMap = map[string]map[string]map[string]string{
	"Aqua": {
		"Youtube": {
			"Id":               "UC1opHUrw8rvnsadT-iGp7Cg",
			"DiscordChannelId": "965968317280055397",
		},
		"Twitch": {
			"Id":               "738746247",
			"DiscordChannelId": "970990916980584508",
		},
		"Twitcasting": {
			"Id":               "1024528894940987392",
			"DiscordChannelId": "969892552486563880",
		},
		"Fanbox": {
			"Id":               "80355000",
			"DiscordChannelId": "965967553870594098",
		},
		"News": {
			"Id":               "湊あくあ",
			"DiscordChannelId": "968838661569400952",
		},
	},
	"Shion": {
		"Youtube": {
			"Id":               "UCXTpFs_3PqI41qX2d9tL2Rw",
			"DiscordChannelId": "965973309432942642",
		},
		"Twitch": {
			"Id":               "773041510",
			"DiscordChannelId": "976847607248871444",
		},
		"Twitcasting": {
			"Id":               "1024533638879166464",
			"DiscordChannelId": "976850346292961310",
		},
		"Fanbox": {
			"Id":               "69014608",
			"DiscordChannelId": "872814425046937610",
		},
		"News": {
			"Id":               "紫咲シオン",
			"DiscordChannelId": "971271623221075968",
		},
	},
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "member",
		Description: "New Member Stream",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "video-id",
				Description: "Youtube Video Id",
				Required:    true,
			},
		},
	},
	{
		Name:        "collab",
		Description: "New Collab Stream",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "collab-with",
				Description: "Which User Is Collabed With",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "Aqua",
						Value: "aqua",
					},
					{
						Name:  "Shion",
						Value: "shion",
					},
					{
						Name:  "Both",
						Value: "both",
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
		Name:        "music",
		Description: "Set Video As Music",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "video-id",
				Description: "Youtube Video Id",
				Required:    true,
			},
		},
	},
	{
		Name:        "add-channels",
		Description: "Add New Channels",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "channel-ids",
				Description: "Youtube Channel Ids (separated by comma)",
				Required:    true,
			},
		},
	},
	{
		Name:        "update-channels-data",
		Description: "Update Channels Data",
	},
	{
		Name:        "get-twitch-access-token",
		Description: "Return Twitch Access Token",
	},
}

var commandsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"member": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if r := recover(); r != nil {
				sendResponse(s, i, "似乎發生了什麼錯誤...")
				log.Println("handler panic:", r)
			}
		}()

		videoId := i.ApplicationCommandData().Options[0].StringValue()

		if !db.Find("Video", "WHERE Id = ?", videoId) {
			video, err := youtube.GetVideo(videoId)
			if err != nil {
				panic(err)
			}

			channel, err := youtube.GetChannel(video.Author.Id)
			if err != nil {
				panic(err)
			}

			err = tools.ImageDownload(video.Thumbnail, "Youtube", channel.Id, "Video", video.Id)
			if err != nil {
				panic(err)
			}

			video.Member = true

			var discordChannelId string

			if video.Author.Id == userDataMap["Aqua"]["Youtube"]["Id"] {
				discordChannelId = userDataMap["Aqua"]["Youtube"]["DiscordChannelId"]
			} else if video.Author.Id == userDataMap["Shion"]["Youtube"]["Id"] {
				discordChannelId = userDataMap["Shion"]["Youtube"]["DiscordChannelId"]
			}

			discord.BaseEmbed("Youtube", channel.Title, channel.Url, channel.Icon).NewNotify("member", video).Send(s, discordChannelId)
			db.Insert("Video", video.Map())
			sendResponse(s, i, "已新增會員限定影片！")
		} else {
			db.Update("Video", videoId, "Member", true)
			sendResponse(s, i, "已將影片改為會員限定！")
		}
	},
	"collab": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if r := recover(); r != nil {
				sendResponse(s, i, "似乎發生了什麼錯誤...")
				log.Println("handler panic:", r)
			}
		}()

		user := i.ApplicationCommandData().Options[0].StringValue()
		videoId := i.ApplicationCommandData().Options[1].StringValue()

		if !db.Find("Video", "WHERE Id = ?", videoId) {
			video, err := youtube.GetVideo(videoId)
			if err != nil {
				panic(err)
			}

			err = tools.ImageDownload(video.Thumbnail, "Youtube", userDataMap[user]["Youtube"]["Id"], "Collab", video.Id)
			if err != nil {
				panic(err)
			}

			discord.BaseEmbed("Youtube", "", "", "").NewNotify("collab", video).Send(s, userDataMap[user]["Youtube"]["DiscordChannelId"])
			db.Insert("Video", video.Map())
			db.Insert("Collab", map[string]any{"VideoId": video.Id, "ChannelId": userDataMap[user]["Youtube"]["Id"]})
			sendResponse(s, i, "已新增連動影片！")
		} else {
			db.Insert("Collab", map[string]any{"VideoId": videoId, "ChannelId": userDataMap[user]["Youtube"]["Id"]})
			sendResponse(s, i, "已新增連動資料！")
		}
	},
	"music": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if r := recover(); r != nil {
				sendResponse(s, i, "似乎發生了什麼錯誤...")
				log.Println("handler panic:", r)
			}
		}()

		videoId := i.ApplicationCommandData().Options[0].StringValue()

		db.Update("Video", videoId, "Music", true)

		sendResponse(s, i, "已將影片設定為音樂")
	},
	"add-channels": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if r := recover(); r != nil {
				sendResponse(s, i, "似乎發生了什麼錯誤...")
				log.Println("handler panic:", r)
			}
		}()

		titles := []string{}
		count := 0

		for _, channelId := range strings.Split(i.ApplicationCommandData().Options[0].StringValue(), ",") {
			channel, err := youtube.GetChannel(channelId)
			if err != nil {
				panic(err)
			}

			db.Insert("Channel", channel.Map())

			titles = append(titles, fmt.Sprintf("***%s***", channel.Title))
			count++
		}

		sendResponse(s, i, fmt.Sprintf("已新增%s共%d筆頻道資料！", strings.Join(titles, "、"), count))
	},
	"update-channels-data": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if r := recover(); r != nil {
				sendResponse(s, i, "似乎發生了什麼錯誤...")
				log.Println("handler panic:", r)
			}
		}()

		for _, channelId := range db.Distinct("channel", "") {
			channel, err := youtube.GetChannel(channelId)
			if err != nil {
				panic(err)
			}

			db.Update("Channel", channelId, "CustomId", channel.CustomId, "Title", channel.Title, "Description", channel.Description,
				"ViewCount", channel.ViewCount, "SubscriberCount", channel.SubscriberCount)
		}

		sendResponse(s, i, "已更新所有頻道資料！")
	},
	"get-twitch-access-token": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		defer func() {
			if r := recover(); r != nil {
				sendResponse(s, i, "似乎發生了什麼錯誤...")
				log.Println("handler panic:", r)
			}
		}()

		token, err := twitch.GetAccessToken()
		if err != nil {
			panic(err)
		}

		sendResponse(s, i, fmt.Sprintf("Twitch Access Token: %s", token))
	},
}

var componentsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, user, videoId string){
	"collab": func(s *discordgo.Session, i *discordgo.InteractionCreate, user, videoId string) {
		defer func() {
			if r := recover(); r != nil {
				sendResponse(s, i, "似乎發生了什麼錯誤...")
				log.Println("handler panic:", r)
			}
		}()

		if !db.Find("Video", "WHERE Id = ?", videoId) {
			video, err := youtube.GetVideo(videoId)
			if err != nil {
				panic(err)
			}

			err = tools.ImageDownload(video.Thumbnail, "Youtube", userDataMap[user]["Youtube"]["Id"], "Collab", video.Id)
			if err != nil {
				panic(err)
			}

			discord.BaseEmbed("Youtube", "", "", "").NewNotify("collab", video).Send(s, userDataMap[user]["Youtube"]["DiscordChannelId"])
			db.Insert("Video", video.Map())
			db.Insert("Collab", map[string]any{"VideoId": video.Id, "ChannelId": userDataMap[user]["Youtube"]["Id"]})
			sendResponse(s, i, "已新增連動影片！")
		} else {
			db.Insert("Collab", map[string]any{"VideoId": videoId, "ChannelId": userDataMap[user]["Youtube"]["Id"]})
			sendResponse(s, i, "已新增連動資料！")
		}
	},
	"no": func(s *discordgo.Session, i *discordgo.InteractionCreate, user, videoId string) {
		defer func() {
			if r := recover(); r != nil {
				sendResponse(s, i, "似乎發生了什麼錯誤...")
				log.Println("handler panic:", r)
			}
		}()

		err := s.ChannelMessageDelete(i.ChannelID, i.Interaction.Message.ID)
		if err != nil {
			panic(err)
		}
	},
}

func sendResponse(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func getComponent(user, videoId string) []discordgo.MessageComponent {
	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					CustomID: fmt.Sprintf("collab:%s:%s", user, videoId),
					Label:    "Yes",
					Emoji:    &discordgo.ComponentEmoji{Name: "✔️"},
					Style:    discordgo.PrimaryButton,
				},
				discordgo.Button{
					CustomID: "no",
					Label:    "No",
					Emoji:    &discordgo.ComponentEmoji{Name: "❌"},
					Style:    discordgo.PrimaryButton,
				},
			},
		},
	}
}

func isContain(list []string, target string) bool {
	for _, element := range list {
		if element == target {
			return true
		}
	}

	return false
}
