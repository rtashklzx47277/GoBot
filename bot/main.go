package main

// history data
// twitter post

import (
	"GoBot/tools"
	"GoBot/tools/discord"
	"GoBot/tools/fanbox"
	"GoBot/tools/sql"
	"GoBot/tools/tiktok"
	"GoBot/tools/twitcasting"
	"GoBot/tools/twitch"
	"GoBot/tools/twitter"
	"GoBot/tools/youtube"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	db            *sql.MySQL
	s             *discordgo.Session
	collabIds     = []string{}
	messageIdList = []string{}
	channelCache  = map[string]*youtube.Channel{}
	logChannelId  = os.Getenv("DISCORD_LOG_CHANNEL_ID")
	testChannelId = os.Getenv("DISCORD_TEST_CHANNEL_ID")

	commands = []*discordgo.ApplicationCommand{
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
	commandsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"collab": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			defer func() {
				if r := recover(); r != nil {
					discord.SendResponse(s, i, "似乎發生了什麼錯誤...")
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

				err = tools.ImageDownload(video.Thumbnail, user, "Youtube", "Collab", video.Id)
				if err != nil {
					panic(err)
				}

				discord.BaseEmbed("Youtube", "", "", "").NewNotify("collab", video).Send(s, tools.UserData[user]["Youtube"]["DiscordChannelId"])
				db.Insert("Video", video.Map())
				db.Insert("Collab", map[string]any{"VideoId": video.Id, "ChannelId": tools.UserData[user]["Youtube"]["Id"]})
				discord.SendResponse(s, i, "已新增連動影片！")
			} else {
				db.Insert("Collab", map[string]any{"VideoId": videoId, "ChannelId": tools.UserData[user]["Youtube"]["Id"]})
				discord.SendResponse(s, i, "已新增連動資料！")
			}
		},
		"music": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			defer func() {
				if r := recover(); r != nil {
					discord.SendResponse(s, i, "似乎發生了什麼錯誤...")
					log.Println("handler panic:", r)
				}
			}()

			videoId := i.ApplicationCommandData().Options[0].StringValue()

			db.Update("Video", videoId, "Music", true)

			discord.SendResponse(s, i, "已將影片設定為音樂")
		},
		"add-channels": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			defer func() {
				if r := recover(); r != nil {
					discord.SendResponse(s, i, "似乎發生了什麼錯誤...")
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

			discord.SendResponse(s, i, fmt.Sprintf("已新增%s共%d筆頻道資料！", strings.Join(titles, "、"), count))
		},
		"update-channels-data": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			defer func() {
				if r := recover(); r != nil {
					discord.SendResponse(s, i, "似乎發生了什麼錯誤...")
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

			discord.SendResponse(s, i, "已更新所有頻道資料！")
		},
		"get-twitch-access-token": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			defer func() {
				if r := recover(); r != nil {
					discord.SendResponse(s, i, "似乎發生了什麼錯誤...")
					log.Println("handler panic:", r)
				}
			}()

			token, err := twitch.GetAccessToken()
			if err != nil {
				panic(err)
			}

			discord.SendResponse(s, i, fmt.Sprintf("Twitch Access Token: %s", token))
		},
	}
	componentsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, user, videoId string){
		"collab": func(s *discordgo.Session, i *discordgo.InteractionCreate, user, videoId string) {
			defer func() {
				if r := recover(); r != nil {
					discord.SendResponse(s, i, "似乎發生了什麼錯誤...")
					log.Println("handler panic:", r)
				}
			}()

			if !db.Find("Video", "WHERE Id = ?", videoId) {
				video, err := youtube.GetVideo(videoId)
				if err != nil {
					panic(err)
				}

				err = tools.ImageDownload(video.Thumbnail, user, "Youtube", "Collab", video.Id)
				if err != nil {
					panic(err)
				}

				discord.BaseEmbed("Youtube", "", "", "").NewNotify("collab", video).Send(s, tools.UserData[user]["Youtube"]["DiscordChannelId"])
				db.Insert("Video", video.Map())
				db.Insert("Collab", map[string]any{"VideoId": video.Id, "ChannelId": tools.UserData[user]["Youtube"]["Id"]})
				discord.SendResponse(s, i, "已新增連動影片！")
			} else {
				db.Insert("Collab", map[string]any{"VideoId": videoId, "ChannelId": tools.UserData[user]["Youtube"]["Id"]})
				discord.SendResponse(s, i, "已新增連動資料！")
			}
		},
		"no": func(s *discordgo.Session, i *discordgo.InteractionCreate, user, videoId string) {
			defer func() {
				if r := recover(); r != nil {
					discord.SendResponse(s, i, "似乎發生了什麼錯誤...")
					log.Println("handler panic:", r)
				}
			}()

			err := s.ChannelMessageDelete(i.ChannelID, i.Interaction.Message.ID)
			if err != nil {
				panic(err)
			}
		},
	}
)

func main() {
	initial()

	getChat("Sakuna", "Roa", "Aqua", "Shion")

	for _, name := range []string{"Sakuna", "Roa", "Aqua", "Shion"} {
		channelId := tools.UserData[name]["Youtube"]["Id"]
		channel, err := youtube.GetChannel(channelId)
		if err != nil {
			panic(err)
		}
		channelCache[channelId] = &youtube.Channel{Title: channel.Title, Url: channel.Url, Icon: channel.Icon}
	}

	runGo(YoutubeStreamNotify, map[string]int{"Sakuna": 30, "Roa": 300, "Aqua": 600, "Shion": 60})
	runGo(YoutubeNotify, map[string]int{"Sakuna": 180, "Roa": 1800, "Aqua": 3600, "Shion": 300})
	runGo(TwitterNotify, map[string]int{"Sakuna": 300, "Roa": 600, "Aqua": 600, "Shion": 600})
	runGo(TweetNotify, map[string]int{"Sakuna": 120})
	runGo(FanboxNotify, map[string]int{"Sakuna": 180, "Roa": 600})
	runGo(Collab, map[string]int{"Shion": 600})

	select {}
}

func YoutubeStreamNotify(name string) {
	defer func() {
		if r := recover(); r != nil {
			tools.DiscordNotify(s, "Youtube Live", name)
			tools.ErrorRecord(r)
		}
	}()

	defer fmt.Printf("%-10s %-20s notification end!\n", name, "Youtube Live")

	channelId, discordChannelId := tools.UserData[name]["Youtube"]["Id"], tools.UserData[name]["Youtube"]["DiscordChannelId"]
	channel := channelCache[channelId]
	baseEmbed := discord.BaseEmbed("Youtube", channel.Title, channel.Url, channel.Icon)

	videoIds := db.Distinct("video", channelId)
	videos, err := youtube.GetPlaylistItems(strings.Replace(channelId, "UC", "UU", 1), 3)
	if err != nil {
		panic(err)
	}

	members, err := youtube.GetPlaylistItems(strings.Replace(channelId, "UC", "UUMO", 1), 3)
	if err != nil {
		panic(err)
	}

	index := len(videos)
	videos = append(videos, members...)

	for i := len(videos) - 1; i >= 0; i-- {
		videoId := videos[i].Id

		if tools.IsContain(videoIds, videoId) {
			continue
		}

		video, err := youtube.GetVideo(videoId)
		if err != nil {
			panic(err)
		}

		err = tools.ImageDownload(video.Thumbnail, name, "Youtube", "Video", video.Id)
		if err != nil {
			panic(err)
		}

		status := ""

		if i >= index {
			video.Member = true
			status = "member"
		} else if video.Live {
			go LiveChat(video.Id, discordChannelId)
		}

		baseEmbed.NewNotify(status, video).Send(s, discordChannelId)
		db.Insert("Video", video.Map())
	}

	oldVideos := db.FindLivestreams(channelId)
	newVideos, err := youtube.GetVideos(db.Distinct("livestream", channelId))
	if err != nil {
		panic(err)
	}

	for i := range oldVideos {
		old, new := oldVideos[i], newVideos[i]

		if new.Private {
			new, err = youtube.GetVideo(new.Id)
			if err != nil {
				panic(err)
			}

			if new.Private {
				thumbnail, err := tools.ImageUpload(old.Thumbnail)
				if err != nil {
					panic(err)
				}

				if old.LiveStatus == 1 {
					baseEmbed.New(old.Title, old.Url, "預定直播已被取消了！", thumbnail).Send(s, discordChannelId)
				} else if old.LiveStatus == 2 {
					baseEmbed.New(old.Title, old.Url, "直播串流已設為不公開！", thumbnail).Send(s, discordChannelId)
				}

				db.Update("Video", old.Id, "Private", new.Private)
			}
		} else if old.ScheduledTime != new.ScheduledTime {
			baseEmbed.New(new.Title, new.Url, "直播預定時間變更了！", new.Thumbnail).CheckAuthor(new.Author.Id).Change(old.ScheduledTime.String("full"), new.ScheduledTime.String("full")).Send(s, discordChannelId)
			db.Update("Video", new.Id, "ScheduledTime", new.ScheduledTime.String())
		} else if old.LiveStatus == 1 && new.LiveStatus == 2 {
			baseEmbed.New(new.Title, new.Url, "直播串流開始了！", new.Thumbnail).CheckAuthor(new.Author.Id).StartTime(new.StartTime).Send(s, discordChannelId)
			db.Update("Video", new.Id, "LiveStatus", new.LiveStatus, "StartTime", new.StartTime.String())
		} else if old.LiveStatus == 2 && new.LiveStatus == 0 {
			baseEmbed.New(new.Title, new.Url, "直播串流結束了！", new.Thumbnail).CheckAuthor(new.Author.Id).EndTime(new.EndTime, new.Length).Send(s, discordChannelId)
			db.Update("Video", new.Id, "LiveStatus", new.LiveStatus, "EndTime", new.EndTime.String(), "Length", new.Length.String())
		}
	}
}

func YoutubeNotify(name string) {
	defer func() {
		if r := recover(); r != nil {
			tools.DiscordNotify(s, "Youtube", name)
			tools.ErrorRecord(r)
		}
	}()

	defer fmt.Printf("%-10s %-20s notification end!\n", name, "Youtube")

	channelId, discordChannelId := tools.UserData[name]["Youtube"]["Id"], tools.UserData[name]["Youtube"]["DiscordChannelId"]
	channelData := db.FindChannel(channelId)
	channel, err := youtube.GetChannel(channelId)
	if err != nil {
		panic(err)
	}

	baseEmbed := discord.BaseEmbed("Youtube", channel.Title, channel.Url, channel.Icon)

	if channelData.SubscriberCount/10000 < channel.SubscriberCount/10000 {
		baseEmbed.New("", "", fmt.Sprintf("Youtube訂閱者數已突破%d萬人了！", channel.SubscriberCount/10000), channel.Icon).Send(s, discordChannelId)
		db.Update("Channel", channelId, "SubscriberCount", channel.SubscriberCount)
		channelData.SubscriberCount = channel.SubscriberCount
	}

	if channelData.ViewCount/50000000 < channel.ViewCount/50000000 {
		baseEmbed.New("", "", fmt.Sprintf("Youtube觀看次數已突破%s億次了！", fmt.Sprintf("%.1f", float64(channel.ViewCount)/100000000)), channel.Icon).Send(s, discordChannelId)
		db.Update("Channel", channelId, "ViewCount", channel.ViewCount)
		channelData.ViewCount = channel.ViewCount
	}

	if channelData.CustomId != channel.CustomId {
		baseEmbed.New("", "", "頻道ID更新了！", "").Change(channelData.CustomId, channel.CustomId).Send(s, discordChannelId)
		db.Update("Channel", channelId, "CustomId", channel.CustomId)
		channelData.CustomId = channel.CustomId
	}

	if channelData.Title != channel.Title {
		baseEmbed.New("", "", "頻道名稱更新了！", "").Change(channelData.Title, channel.Title).Send(s, discordChannelId)
		db.Update("Channel", channelId, "Title", channel.Title)
		channelData.Title = channel.Title
	}

	if channelData.Description != channel.Description {
		baseEmbed.New("", "", "頻道介紹更新了！", "").Change(channelData.Description, channel.Description).Send(s, discordChannelId)
		db.Update("Channel", channelId, "Description", channel.Description)
		channelData.Description = channel.Description
	}

	if check, image, err := tools.ImageCheck(channelData.Icon, channel.Icon); err == nil && check == 0 {
		err = tools.ImageDownload(channel.Icon, name, "Youtube", "Icon")
		if err != nil {
			panic(err)
		}
		channelData.Icon = channel.Icon

		baseEmbed.New("", "", "頻道頭貼更新了！", image).Send(s, discordChannelId)
	} else if err != nil {
		panic(err)
	}

	if check, image, err := tools.ImageCheck(channelData.Banner, channel.Banner); err == nil && check == 0 {
		err = tools.ImageDownload(channel.Banner, name, "Youtube", "Banner")
		if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "頻道橫幅更新了！", image).Send(s, discordChannelId)
	} else if err != nil {
		panic(err)
	}

	oldVideos := db.FindVideos(channelId)
	newVideos, err := youtube.GetVideos(db.Distinct("video", channelId))
	if err != nil {
		panic(err)
	}

	for i := range oldVideos {
		old, new := oldVideos[i], newVideos[i]

		if !old.Private && new.Private {
			new, err = youtube.GetVideo(new.Id)
			if err != nil {
				panic(err)
			}

			if new.Private {
				thumbnail, err := tools.ImageUpload(old.Thumbnail)
				if err != nil {
					panic(err)
				}

				baseEmbed.New(old.Title, old.Url, "影片已設為不公開了！", thumbnail).Send(s, discordChannelId)
				db.Update("Video", old.Id, "Private", new.Private)
			}
		} else if old.Private && !new.Private {
			// turn to public
			err = tools.ImageDownload(new.Thumbnail, name, "Youtube", "Video", new.Id)
			if err != nil {
				panic(err)
			}

			commentIds := db.Distinct("comment", channelId)
			comments, err := youtube.GetComments("video", new.Id)
			if err != nil {
				panic(err)
			}

			for _, comment := range comments {
				if !tools.IsContain(commentIds, comment.Id) {
					db.Insert("Comment", comment.Map())
				}
			}
			// turn to public

			baseEmbed.New(new.Title, new.Url, "影片已設為公開了！", new.Thumbnail).Send(s, discordChannelId)
			db.Update("Video", new.Id, "Title", new.Title, "Description", new.Description, "Length", new.Length.String(), "ViewCount", new.ViewCount, "LiveStatus", new.LiveStatus,
				"PublishedTime", new.PublishedTime.String(), "ScheduledTime", new.ScheduledTime.String(), "StartTime", new.StartTime.String(), "EndTime", new.EndTime.String(),
				"Comment", new.Comment, "Live", new.Live, "Private", new.Private)
		} else if !old.Private && !new.Private {
			if old.Comment && !new.Comment {
				baseEmbed.New(new.Title, new.Url, "影片留言功能已停用！", new.Thumbnail).Send(s, discordChannelId)
				db.Update("Video", new.Id, "Comment", new.Comment)
			} else if !old.Comment && new.Comment {
				baseEmbed.New(new.Title, new.Url, "影片留言功能已啟用！", new.Thumbnail).Send(s, discordChannelId)
				db.Update("Video", new.Id, "Comment", new.Comment)
			}

			if old.Title != new.Title {
				description := "直播串流標題變更了！"

				if new.LiveStatus == 0 {
					description = "影片標題更新了！"
				}

				baseEmbed.New(new.Title, new.Url, description, new.Thumbnail).Change(old.Title, new.Title).Send(s, discordChannelId)
				db.Update("Video", new.Id, "Title", new.Title)
			}

			if old.Description != new.Description {
				description := "直播串流資訊欄更新了！"

				if new.LiveStatus == 0 {
					description = "影片資訊欄更新了！"
				}

				baseEmbed.New(new.Title, new.Url, description, new.Thumbnail).Change(old.Description, new.Description).Send(s, testChannelId)
				db.Update("Video", new.Id, "Description", new.Description)
			}

			if old.Length != new.Length {
				if old.Length != tools.Duration(0) {
					baseEmbed.New(new.Title, new.Url, "影片長度更新了！", new.Thumbnail).Change(old.Length.String(), new.Length.String()).Send(s, testChannelId)
				}

				db.Update("Video", new.Id, "Length", new.Length.String())
			}

			if old.PublishedTime != new.PublishedTime {
				db.Update("Video", new.Id, "PublishedTime", new.PublishedTime.String())
			}

			// if old.StartTime != new.StartTime {
			// 	baseEmbed.New(new.Title, new.Url, "影片開始時間更新了！", new.Thumbnail).Change(old.StartTime.String("full"), new.StartTime.String("full")).Send(s, testChannelId)
			// 	db.Update("Video", new.Id, "StartTime", new.StartTime.String())
			// }

			// if old.EndTime != new.EndTime {
			// 	baseEmbed.New(new.Title, new.Url, "影片結束時間更新了！", new.Thumbnail).Change(old.EndTime.String("full"), new.EndTime.String("full")).Send(s, testChannelId)
			// 	db.Update("Video", new.Id, "EndTime", new.EndTime.String())
			// }

			if old.Music && ((new.ViewCount < 1000000 && new.ViewCount/100000 > old.ViewCount/100000) || (new.ViewCount >= 1000000 && new.ViewCount/500000 > old.ViewCount/500000)) {
				baseEmbed.New(new.Title, new.Url, fmt.Sprintf("影片觀看次數已突破%d萬次了！", new.ViewCount/10000), new.Thumbnail).Send(s, discordChannelId)
				db.Update("Video", new.Id, "ViewCount", new.ViewCount)
			} else if !old.Music && new.ViewCount/100000 > old.ViewCount/100000 {
				baseEmbed.New(new.Title, new.Url, fmt.Sprintf("影片觀看次數已突破%d萬次了！", new.ViewCount/10000), new.Thumbnail).Send(s, testChannelId)
				db.Update("Video", new.Id, "ViewCount", new.ViewCount)
			}

			if new.LiveStatus != 0 {
				check, image, err := tools.ImageCheck(old.Thumbnail, new.Thumbnail)
				if err != nil {
					panic(err)
				}

				if check != 1 {
					err = tools.ImageDownload(new.Thumbnail, name, "Youtube", "Video", new.Id)
					if err != nil {
						panic(err)
					}

					if check == 0 {
						baseEmbed.New(new.Title, new.Url, "直播串流封面更新了！", image).Send(s, discordChannelId)
					}
				}
			}
		}
	}

	videoIds := db.Distinct("public", channelId)
	videos, err := youtube.GetPlaylistItems(strings.Replace(channelId, "UC", "UUMO", 1), 50)
	if err != nil {
		panic(err)
	}

	for _, video := range videos {
		if tools.IsContain(videoIds, video.Id) {
			baseEmbed.New(video.Title, video.Url, "影片已從公開轉為會員限定了！", video.Thumbnail).Send(s, discordChannelId)
			db.Update("Video", video.Id, "Member", true)
		}
	}

	videoIds = db.Distinct("member", channelId)
	videos, err = youtube.GetPlaylistItems(strings.Replace(channelId, "UC", "UU", 1), 50)
	if err != nil {
		panic(err)
	}

	for _, video := range videos {
		if tools.IsContain(videoIds, video.Id) {
			baseEmbed.New(video.Title, video.Url, "影片已從會員限定轉為公開了！", video.Thumbnail).Send(s, discordChannelId)
			db.Update("Video", video.Id, "Member", false)
		}
	}

	oldPlaylists := db.FindPlaylists(channelId)
	newPlaylists, err := youtube.GetPlaylists(channelId)
	if err != nil {
		panic(err)
	}

	for _, playlists := range youtube.GroupPlaylist(oldPlaylists, newPlaylists) {
		if playlists.New == nil {
			image, err := tools.ImageUpload(playlists.Old.Thumbnail)
			if err != nil {
				panic(err)
			}

			baseEmbed.New(playlists.Old.Title, playlists.Old.Url, "播放清單已被刪除！", image).Send(s, discordChannelId)
			db.Delete("PlaylistItem", "WHERE PlaylistId = ?", playlists.Old.Id)
			db.Delete("Playlist", "WHERE Id = ?", playlists.Old.Id)
			tools.ImageRemove(playlists.Old.Thumbnail)
		} else if playlists.Old == nil {
			err = tools.ImageDownload(playlists.New.Thumbnail, name, "Youtube", "Playlist", playlists.New.Id)
			if err != nil {
				panic(err)
			}

			playlistItems, err := youtube.GetPlaylistItems(playlists.New.Id, 50)
			if err != nil {
				panic(err)
			}

			baseEmbed.New(playlists.New.Title, playlists.New.Url, "建立了新的播放清單！", playlists.New.Thumbnail).Send(s, discordChannelId)
			db.Insert("Playlist", playlists.New.Map())

			for _, playlistItem := range playlistItems {
				db.Insert("PlaylistItem", map[string]any{"PlaylistId": playlists.New.Id, "VideoId": playlistItem.Id})
			}
		} else {
			if playlists.Old.Title != playlists.New.Title {
				baseEmbed.New(playlists.New.Title, playlists.New.Url, "播放清單名稱更新了！", playlists.New.Thumbnail).Change(playlists.Old.Title, playlists.New.Title).Send(s, discordChannelId)
				db.Update("Playlist", playlists.New.Id, "Title", playlists.New.Title)
			}

			if playlists.Old.Description != playlists.New.Description {
				baseEmbed.New(playlists.New.Title, playlists.New.Url, "播放清單資訊欄更新了！", playlists.New.Thumbnail).Change(playlists.Old.Description, playlists.New.Description).Send(s, discordChannelId)
				db.Update("Playlist", playlists.New.Id, "Description", playlists.New.Description)
			}

			if check, image, err := tools.ImageCheck(playlists.Old.Thumbnail, playlists.New.Thumbnail); err == nil && check == 0 {
				err = tools.ImageDownload(playlists.New.Thumbnail, name, "Youtube", "Playlist", playlists.New.Id)
				if err != nil {
					panic(err)
				}

				baseEmbed.New(playlists.New.Title, playlists.New.Url, "播放清單封面更新了！", image).Send(s, discordChannelId)
			} else if err != nil {
				panic(err)
			}

			oldPlaylistItems := db.FindPlaylistItems(playlists.New.Id)
			newPlaylistItems, err := youtube.GetPlaylistItems(playlists.New.Id, 50)
			if err != nil {
				panic(err)
			}

			for _, playlistItems := range youtube.GroupVideo(oldPlaylistItems, newPlaylistItems) {
				if playlistItems.New == nil {
					image, err := tools.ImageUpload(playlistItems.Old.Thumbnail)
					if err != nil {
						panic(err)
					}

					baseEmbed.New(playlistItems.Old.Title, playlistItems.Old.Url, fmt.Sprintf("自「%s」中移除了影片！", playlists.New.Title), image).Send(s, discordChannelId)
					db.Delete("PlaylistItem", "WHERE PlaylistId = ? AND VideoId = ?", playlists.New.Id, playlistItems.Old.Id)
				} else if playlistItems.Old == nil {
					if !db.Find("Video", "WHERE Id = ?", playlistItems.New.Id) {
						video, err := youtube.GetVideo(playlistItems.New.Id)
						if err != nil {
							panic(err)
						}

						if video.Author.Id == "" {
							video.Author.Id = channelId
						}

						db.Insert("Video", video.Map())
					}

					baseEmbed.New(playlistItems.New.Title, playlistItems.New.Url, fmt.Sprintf("追加新影片至「%s」中！", playlists.New.Title), playlistItems.New.Thumbnail).Send(s, discordChannelId)
					db.Insert("PlaylistItem", map[string]any{"PlaylistId": playlists.New.Id, "VideoId": playlistItems.New.Id})
				}
			}
		}
	}

	commentIds := db.Distinct("comment", channelId)
	replyIds := db.Distinct("reply", channelId)
	comments, err := youtube.GetComments("channel", channelId)
	if err != nil {
		panic(err)
	}

	for _, comment := range comments {
		if !tools.IsContain(commentIds, comment.Id) {
			comment = db.CompelteComment(comment)
			s.ChannelMessageSend(testChannelId, fmt.Sprintf("「[%s](<%s>)」在「[%s](<%s>)」的影片「[%s](<%s>)」中發表留言：\n> %s",
				db.FindChannelTitle(comment.Author.Id), comment.Author.Url, channel.Title, channel.Url, db.FindVideoTitle(comment.Video.Id), comment.Video.Url, strings.Replace(comment.Text, "\n", "\n> ", -1)))
			db.Insert("Comment", comment.Map())
		}

		replies, err := youtube.GetComments("reply", comment.Id)
		if err != nil {
			panic(err)
		}

		for _, reply := range replies {
			if !tools.IsContain(replyIds, reply.Id) {
				reply = db.CompelteComment(reply)
				s.ChannelMessageSend(testChannelId, fmt.Sprintf("「[%s](<%s>)」在「[%s](<%s>)」的影片「[%s](<%s>)」中發表留言：\n> %s",
					db.FindChannelTitle(reply.Author.Id), reply.Author.Url, channel.Title, channel.Url, db.FindVideoTitle(reply.Video.Id), reply.Video.Url, strings.Replace(reply.Text, "\n", "\n> ", -1)))
				db.Insert("Comment", reply.Map())
			}
		}
	}

	postIds := db.Distinct("post", channelId)
	posts, err := youtube.GetCommunity(channelId)
	if err != nil {
		panic(err)
	}

	for _, post := range posts {
		if !tools.IsContain(postIds, post.Id) {
			var description string

			if post.Member {
				description = "會員限定"
			}

			if post.Renderer.Type == "Image" {
				for i, image := range post.Renderer.Images {
					err = tools.ImageDownload(image, name, "Youtube", "Post", fmt.Sprintf("%s_%d", post.Id, i+1))
					if err != nil {
						panic(err)
					}
				}
			}

			s.ChannelMessageSend(discordChannelId, fmt.Sprintf("有新的%s社群投稿！ <%s>", description, post.Url))
			db.Insert("Post", post.Map())

			if post.Renderer.Type == "Poll" || post.Renderer.Type == "Quiz" {
				for _, choice := range post.Renderer.Choices {
					choiceMap := choice.Map()
					choiceMap["PostId"] = post.Id
					db.Insert("Choice", choiceMap)
				}
			}
		}
	}

	if name == "Aqua" || name == "Roa" {
		return
	}

	badgeIds := db.Distinct("badge", channelId)
	stampIds := db.Distinct("stamp", channelId)
	oldPerks := db.FindPerks(channelId)
	badges, stamps, newPerks, err := youtube.GetMemberShip(channelId)
	if errors.Is(err, youtube.ErrorNoMembership) {
		s.ChannelMessageSend(logChannelId, fmt.Sprintf("%s Youtube會員Cookie可能已過期！", name))
		return
	} else if err != nil {
		panic(err)
	}

	for _, badge := range badges {
		if !tools.IsContain(badgeIds, badge.Label) {
			err := tools.ImageDownload(badge.Image, name, "Youtube", "Badge", badge.Label)
			if err != nil {
				panic(err)
			}

			baseEmbed.New("", "", fmt.Sprintf("新增了第 %s 個月的徽章", badge.Label), badge.Image).Send(s, discordChannelId)
			db.Insert("Badge", badge.Map())
		} else {
			if check, image, err := tools.ImageCheck(fmt.Sprintf("/bot/media/%s/Youtube/Badge/%s.jpg", name, badge.Label), badge.Image); err == nil && check == 0 {
				err = tools.ImageDownload(badge.Image, name, "Youtube", "Badge", badge.Label)
				if err != nil {
					panic(err)
				}

				baseEmbed.New("", "", fmt.Sprintf("第 %s 個月的徽章更新了！", badge.Label), image).Send(s, discordChannelId)
			} else if err != nil {
				panic(err)
			}
		}
	}

	for _, stamp := range stamps {
		if !tools.IsContain(stampIds, stamp.Label) {
			err := tools.ImageDownload(stamp.Image, name, "Youtube", "Stamp", stamp.Label)
			if err != nil {
				panic(err)
			}

			baseEmbed.New("", "", "新增了自訂表情符號", stamp.Image).Send(s, discordChannelId)
			db.Insert("Stamp", stamp.Map())
		} // else {
		// 	if check, image, err := tools.ImageCheck(fmt.Sprintf("/bot/media/%s/Youtube/Stamp/%s.jpg", name, stamp.Label), stamp.Image); err == nil && check == 0 {
		// 		err = tools.ImageDownload(stamp.Image, name, "Youtube", "Stamp", stamp.Label)
		// 		if err != nil {
		// 			panic(err)
		// 		}

		// 		baseEmbed.New("", "", "自訂表情符號更新了！", image).Send(s, discordChannelId)
		// 	} else if err != nil {
		// 		panic(err)
		// 	}
		// }
	}

	for _, perks := range youtube.GroupPerk(oldPerks, newPerks) {
		if perks.New == nil {
			baseEmbed.New(perks.Old.Title, "", "會員福利已被刪除！", "").Send(s, discordChannelId)
			db.Delete("Perk", "WHERE ChannelId = ? AND Title = ?", channelId, perks.Old.Title)
		} else if perks.Old == nil {
			baseEmbed.New(perks.New.Title, "", "新增了會員福利！", "").Send(s, discordChannelId)
			db.Insert("Perk", perks.New.Map())
		}
	}
}

func Collab(name string) {
	defer func() {
		if r := recover(); r != nil {
			tools.DiscordNotify(s, "Collab", name)
			tools.ErrorRecord(r)
		}
	}()

	defer fmt.Printf("%-10s %-20s notification end!\n", name, "Collab")

	channelId, discordChannelId := tools.UserData[name]["Youtube"]["Id"], tools.UserData[name]["Youtube"]["DiscordChannelId"]

	oldIds := db.Distinct("collab", channelId)
	newIds, err := youtube.GetCollab(channelId)
	if err != nil {
		panic(err)
	}

	for _, videoId := range newIds {
		if tools.IsContain(oldIds, videoId) {
			continue
		}

		video, err := youtube.GetVideo(videoId)
		if err != nil {
			panic(err)
		}

		err = tools.ImageDownload(video.Thumbnail, name, "Youtube", "Collab", video.Id)
		if err != nil {
			panic(err)
		}

		discord.BaseEmbed("Youtube", "", "", "").NewNotify("collab", video).Send(s, discordChannelId)
		db.Insert("Video", video.Map())
		db.Insert("Collab", map[string]any{"VideoId": video.Id, "ChannelId": channelId})
	}

	oldIds = db.Distinct("collab", channelId)
	videos, err := youtube.GetSchedule(name)
	if err != nil {
		panic(err)
	}

	for _, video := range videos {
		if tools.IsContain(collabIds, video.Id) || tools.IsContain(oldIds, video.Id) {
			continue
		}

		s.ChannelMessageSendComplex(testChannelId, &discordgo.MessageSend{
			Embed: (*discordgo.MessageEmbed)(discord.BaseEmbed("Youtube", "", "", "").NewNotify("collab", video)),
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							CustomID: fmt.Sprintf("collab:%s:%s", name, video.Id),
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
			},
		})

		collabIds = append(collabIds, video.Id)
	}
}

func TwitterNotify(name string) {
	defer func() {
		if r := recover(); r != nil {
			tools.DiscordNotify(s, "Twitter", name)
			tools.ErrorRecord(r)
		}
	}()

	defer fmt.Printf("%-10s %-20s notification end!\n", name, "Twitter")

	userId, username, discordChannelId := tools.UserData[name]["Twitter"]["Id"], tools.UserData[name]["Twitter"]["Username"], testChannelId // tools.UserData[name]["Twitter"]["DiscordChannelId"]

	userData := db.FindTwitterUser(userId)
	user, err := twitter.GetUser(username)
	if errors.Is(err, tools.ErrorTooManyRequests) {
		fmt.Println("Out of quota! Please wait... ")
		return
	} else if err != nil {
		panic(err)
	}

	baseEmbed := discord.BaseEmbed("Twitter", user.Name, user.Url, user.Icon)

	if userId != user.Id {
		newUsername, err := twitter.GetUsername(userId)
		if errors.Is(err, tools.ErrorTooManyRequests) {
			fmt.Println("Out of quota! Please wait... ")
			return
		} else if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "用戶Id更新了！", "").Change(username, newUsername).Send(s, discordChannelId)
		db.Update("TwitterUser", userId, "Username", newUsername)
		tools.UserData[name]["Twitter"]["Username"] = newUsername

		return
	}

	if userData.Name != user.Name {
		baseEmbed.New("", "", "用戶名稱更新了！", "").Change(userData.Name, user.Name).Send(s, discordChannelId)
		db.Update("TwitterUser", userId, "Name", user.Name)
	}

	if userData.Description != user.Description {
		baseEmbed.New("", "", "用戶介紹欄更新了！", "").Change(userData.Description, user.Description).Send(s, discordChannelId)
		db.Update("TwitterUser", userId, "Description", user.Description)
	}

	if userData.Location != user.Location {
		baseEmbed.New("", "", "用戶位置更新了！", "").Change(userData.Location, user.Location).Send(s, discordChannelId)
		db.Update("TwitterUser", userId, "Location", user.Location)
	}

	if userData.Link != user.Link {
		baseEmbed.New("", "", "用戶連結更新了！", "").Change(userData.Link, user.Link).Send(s, discordChannelId)
		db.Update("TwitterUser", userId, "Link", user.Link)
	}

	if userData.Pinned != user.Pinned {
		var message string

		if user.Pinned == "" {
			message = fmt.Sprintf("取消釘選了推文！ https://x.com/x/status/%s", userData.Pinned)
		} else {
			message = fmt.Sprintf("釘選了新的推文！ https://x.com/x/status/%s", user.Pinned)
		}

		s.ChannelMessageSend(discordChannelId, message)
		db.Update("TwitterUser", userId, "Pinned", user.Pinned)
	}

	if userData.FollowersCount/10000 < user.FollowersCount/10000 {
		baseEmbed.New("", "", fmt.Sprintf("Twitter追隨者數已突破%d萬人了！", user.FollowersCount/10000), user.Icon).Send(s, discordChannelId)
		db.Update("TwitterUser", userId, "FollowersCount", user.FollowersCount)
	}

	if userData.FollowingCount != user.FollowingCount && user.FollowingCount != 0 {
		var message string

		if userData.FollowingCount < user.FollowingCount {
			message = "追隨了新的用戶！"
		} else {
			message = "解除追隨了用戶！"
		}

		s.ChannelMessageSend(discordChannelId, fmt.Sprintf("%s %s FollowingCount: %d -> %d", user.Name, message, userData.FollowingCount, user.FollowingCount))
		db.Update("TwitterUser", userId, "FollowingCount", user.FollowingCount)
	}

	if userData.LikeCount < user.LikeCount {
		s.ChannelMessageSend(discordChannelId, fmt.Sprintf("%s 對新的推文點了喜歡！ LikeCount: %d -> %d", user.Name, userData.LikeCount, user.LikeCount))
		db.Update("TwitterUser", userId, "LikeCount", user.LikeCount)
	}

	if check, image, err := tools.ImageCheck(userData.Icon, user.Icon); err == nil && check == 0 {
		err = tools.ImageDownload(user.Icon, name, "Twitter", "Icon")
		if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "用戶頭貼更新了！", image).Send(s, discordChannelId)
	} else if err != nil {
		panic(err)
	}

	if check, image, err := tools.ImageCheck(userData.Banner, user.Banner); err == nil && check == 0 {
		err = tools.ImageDownload(user.Banner, name, "Twitter", "Banner")
		if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "用戶橫幅更新了！", image).Send(s, discordChannelId)
	} else if err != nil {
		panic(err)
	}
}

func TweetNotify(name string) {
	defer func() {
		if r := recover(); r != nil {
			tools.DiscordNotify(s, "Tweet", name)
			tools.ErrorRecord(r)
		}
	}()

	defer fmt.Printf("%-10s %-20s notification end!\n", name, "Tweet")

	userId, discordChannelId := tools.UserData[name]["Twitter"]["Id"], testChannelId // tools.UserData[name]["Twitter"]["DiscordChannelId"]
	user := db.FindTwitterUser(userId)

	posts, err := twitter.GetTimeline(userId, user.Latest)
	if errors.Is(err, tools.ErrorTooManyRequests) {
		fmt.Println("Out of tweet quota! Please wait... ")
		return
	} else if err != nil {
		panic(err)
	}

	for i := len(posts) - 1; i >= 0; i-- {
		post := posts[i]

		for i, media := range post.Media {
			if media.Type == "photo" {
				err := tools.ImageDownload(media.Url, name, "Twitter", "Post", fmt.Sprintf("%s_%d", post.Id, i+1))
				if err != nil {
					panic(err)
				}
			} else {
				fmt.Println(media.Url)
				err := tools.VideoDownload(media.Url, name, "Twitter", "Post", fmt.Sprintf("%s_%d", post.Id, i+1))
				if err != nil {
					panic(err)
				}
			}
		}

		if post.PollId != "" {
			for _, option := range post.Options {
				db.Insert("TwitterPoll", map[string]any{"Id": post.PollId, "Option": option})
			}
		}

		var message string
		if post.IsRetweeted {
			message = "轉推了一則推文！"
		} else if post.IsReplied {
			message = "回覆了一則推文！"
		} else {
			message = "發布了新的推文！"
		}

		s.ChannelMessageSend(discordChannelId, fmt.Sprintf("%s(@%s) %s %s", user.Name, user.Username, message, post.Url))
		db.Insert("TwitterPost", post.Map())
		db.Update("TwitterUser", userId, "Latest", post.Id)
	}
}

func TwitchStreamNotify(name string) {
	defer func() {
		if r := recover(); r != nil {
			tools.DiscordNotify(s, "Twitch Live", name)
			tools.ErrorRecord(r)
		}
	}()

	defer fmt.Printf("%-10s %-20s notification end!\n", name, "Twitch Live")

	userId, discordChannelId := tools.UserData[name]["Twitch"]["Id"], tools.UserData[name]["Twitch"]["DiscordChannelId"]

	userData := db.FindTwitchUser(userId)
	user, err := twitch.GetUser(userId)
	if err != nil {
		panic(err)
	}

	baseEmbed := discord.BaseEmbed("Twitch", user.Title, user.Url, user.Icon)

	live, title, err := twitch.GetLive(userId)
	if err != nil {
		panic(err)
	}

	if !userData.Live && live {
		baseEmbed.New(title, user.Url, "直播串流開始了！", user.Thumbnail).Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "Live", 1)
	} else if userData.Live && !live {
		baseEmbed.New("", "", "直播串流結束了！", user.Thumbnail).Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "Live", 0)
	}
}

func TwitchNotify(name string) {
	defer func() {
		if r := recover(); r != nil {
			tools.DiscordNotify(s, "Twitch", name)
			tools.ErrorRecord(r)
		}
	}()

	defer fmt.Printf("%-10s %-20s notification end!\n", name, "Twitch")

	userId, discordChannelId := tools.UserData[name]["Twitch"]["Id"], tools.UserData[name]["Twitch"]["DiscordChannelId"]

	userData := db.FindTwitchUser(userId)
	user, err := twitch.GetUser(userId)
	if err != nil {
		panic(err)
	}

	baseEmbed := discord.BaseEmbed("Twitch", user.Title, user.Url, user.Icon)

	schedules, err := twitch.GetHoloSchedule(name)
	if err != nil {
		panic(err)
	}

	for _, schedule := range schedules {
		if !db.Find("TwitchSchedule", "WHERE UserId = ? AND ScheduledTime = ?", userId, schedule.ScheduledTime.String()) {
			baseEmbed.New("", "", "有新的直播預定！", user.Thumbnail).UpcomingTime(schedule.ScheduledTime).Send(s, discordChannelId)
			db.Insert("TwitchSchedule", map[string]any{"UserId": userId, "ScheduledTime": schedule.ScheduledTime.String()})
		}
	}

	// if name == "Aqua" {
	// 	schedules, err = twitch.GetSchedule(userId)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	for _, schedule := range schedules {
	// 		if !db.Find("TwitchSchedule", "WHERE UserId = ? AND ScheduledTime = ?", userId, schedule.ScheduledTime.String()) {
	// 			baseEmbed.New("", "", "有新的直播預定！", user.Thumbnail).UpcomingTime(schedule.ScheduledTime).Send(s, discordChannelId)
	// 			db.Insert("TwitchSchedule", map[string]any{"UserId": userId, "ScheduledTime": schedule.ScheduledTime.String()})
	// 		}
	// 	}
	// }

	if userData.LoginId != user.LoginId {
		baseEmbed.New("", "", "用戶ID更新了！", "").Change(userData.LoginId, user.LoginId).Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "LoginId", user.LoginId)
	}

	if userData.Title != user.Title {
		baseEmbed.New("", "", "用戶名稱更新了！", "").Change(userData.Title, user.Title).Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "Title", user.Title)
	}

	if userData.Description != user.Description {
		baseEmbed.New("", "", "用戶介紹更新了！", "").Change(userData.Description, user.Description).Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "Description", user.Description)
	}

	if userData.ChannelTitle != user.ChannelTitle {
		baseEmbed.New("", "", "用戶頻道標題更新了！", "").Change(userData.ChannelTitle, user.ChannelTitle).Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "ChannelTitle", user.ChannelTitle)
	}

	if userData.Color != user.Color {
		baseEmbed.New("", "", "用戶頻道聊天室名稱色彩更新了！", "").Change(userData.Color, user.Color).Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "Color", user.Color)
	}

	if !userData.EmoteMode && user.EmoteMode {
		baseEmbed.New("", "", "已啟用聊天室表情符號限定模式了！", "").Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "EmoteMode", user.EmoteMode)
	} else if userData.EmoteMode && !user.EmoteMode {
		baseEmbed.New("", "", "已解除聊天室表情符號限定模式了！", "").Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "EmoteMode", user.EmoteMode)
	}

	if !userData.SubscriberMode && user.SubscriberMode {
		baseEmbed.New("", "", "已啟用聊天室訂閱者限定模式了！", "").Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "SubscriberMode", user.SubscriberMode)
	} else if userData.SubscriberMode && !user.SubscriberMode {
		baseEmbed.New("", "", "已解除聊天室訂閱者限定模式了！", "").Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "SubscriberMode", user.SubscriberMode)
	}

	if !userData.SubscriberMode && user.SubscriberMode {
		baseEmbed.New("", "", "已啟用聊天室訂閱者限定模式了！", "").Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "SubscriberMode", user.SubscriberMode)
	} else if userData.SubscriberMode && !user.SubscriberMode {
		baseEmbed.New("", "", "已解除聊天室訂閱者限定模式了！", "").Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "SubscriberMode", user.SubscriberMode)
	}

	if !userData.UniqueMode && user.UniqueMode {
		baseEmbed.New("", "", "已啟用聊天室不重複聊天模式了！", "").Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "UniqueMode", user.UniqueMode)
	} else if userData.UniqueMode && !user.UniqueMode {
		baseEmbed.New("", "", "已解除聊天室不重複聊天模式了！", "").Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "UniqueMode", user.UniqueMode)
	}

	if !userData.FollowMode && user.FollowMode {
		baseEmbed.New("", "", fmt.Sprintf("已啟用聊天室追隨者限定模式了！ (%d分鐘)", user.FollowTime), "").Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "FollowMode", user.FollowMode, "FollowTime", user.FollowTime)
	} else if userData.FollowMode && !user.FollowMode {
		baseEmbed.New("", "", "已解除聊天室追隨者限定模式了！", "").Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "FollowMode", user.FollowMode)
	} else if userData.FollowMode && user.FollowMode && userData.FollowTime != user.FollowTime {
		baseEmbed.New("", "", fmt.Sprintf("已將追隨者限定模式設定為%d分鐘了！", user.FollowTime), "").Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "FollowTime", user.FollowTime)
	}

	if !userData.SlowMode && user.SlowMode {
		baseEmbed.New("", "", fmt.Sprintf("已啟用聊天室發言時間限制模式了！ (%d秒鐘)", user.SlowTime), "").Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "SlowMode", user.SlowMode, "SlowTime", user.SlowTime)
	} else if userData.SlowMode && !user.SlowMode {
		baseEmbed.New("", "", "已解除聊天室發言時間限制模式了！", "").Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "SlowMode", user.SlowMode)
	} else if userData.SlowMode && user.SlowMode && userData.SlowTime != user.SlowTime {
		baseEmbed.New("", "", fmt.Sprintf("已將發言時間限制模式設定為%d秒鐘了！", user.SlowTime), "").Send(s, discordChannelId)
		db.Update("TwitchUser", userId, "SlowTime", user.SlowTime)
	}

	if check, image, err := tools.ImageCheck(userData.Icon, user.Icon); err == nil && check == 0 {
		err = tools.ImageDownload(user.Icon, name, "Twitch", "Icon")
		if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "用戶頻道頭貼更新了！", image).Send(s, discordChannelId)
	} else if err != nil {
		panic(err)
	}

	if check, image, err := tools.ImageCheck(userData.Thumbnail, user.Thumbnail); err == nil && check == 0 {
		err = tools.ImageDownload(user.Thumbnail, name, "Twitch", "Thumbnail")
		if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "用戶頻道待機圖更新了！", image).Send(s, discordChannelId)
	} else if err != nil {
		panic(err)
	}

	oldBadges := db.FindTwitchBadges(userId)
	newBadges, err := twitch.GetBadges(userId)
	if err != nil {
		panic(err)
	}

	for _, badges := range twitch.GroupBadge(oldBadges, newBadges) {
		if badges.New == nil {
			image, err := tools.ImageUpload(badges.Old.Image)
			if err != nil {
				panic(err)
			}

			baseEmbed.New(badges.Old.Title, "", "用戶頻道徽章已被刪除！", image).Send(s, discordChannelId)
			db.Delete("TwitchBadge", "WHERE Id = ?", badges.Old.Id)
			tools.ImageRemove(badges.Old.Image)
		} else if badges.Old == nil {
			err = tools.ImageDownload(badges.New.Image, name, "Twitch", "Badge", badges.New.Id)
			if err != nil {
				panic(err)
			}

			baseEmbed.New(badges.New.Title, "", "新增了新的用戶頻道徽章！", badges.New.Image).Send(s, discordChannelId)
			db.Insert("TwitchBadge", badges.New.Map())
		}
	}

	oldStamps := db.FindTwitchStamps(userId)
	newStamps, err := twitch.GetStamps(userId)
	if err != nil {
		panic(err)
	}

	for _, stamps := range twitch.GroupStamp(oldStamps, newStamps) {
		if stamps.New == nil {
			image, err := tools.ImageUpload(stamps.Old.Image)
			if err != nil {
				panic(err)
			}

			baseEmbed.New(stamps.Old.Title, "", "用戶頻道貼圖已被刪除！", image).Send(s, discordChannelId)
			db.Delete("TwitchStamp", "WHERE Id = ?", stamps.Old.Id)
			tools.ImageRemove(stamps.Old.Image)
		} else if stamps.Old == nil {
			err = tools.ImageDownload(stamps.New.Image, name, "Twitch", "Stamp", stamps.New.Id)
			if err != nil {
				panic(err)
			}

			baseEmbed.New(stamps.New.Title, "", "新增了新的用戶頻道貼圖！", stamps.New.Image).Send(s, discordChannelId)
			db.Insert("TwitchStamp", stamps.New.Map())
		}
	}
}

func TwitcastingStreamNotify(name string) {
	defer func() {
		if r := recover(); r != nil {
			tools.DiscordNotify(s, "Twitcasting Live", name)
			tools.ErrorRecord(r)
		}
	}()

	defer fmt.Printf("%-10s %-20s notification end!\n", name, "Twitcasting Live")

	userId, discordChannelId := tools.UserData[name]["Twitcasting"]["Id"], tools.UserData[name]["Twitcasting"]["DiscordChannelId"]
	user := db.FindTwitcastingUser(userId)

	live, title, err := twitcasting.GetStream(userId)
	if err != nil {
		panic(err)
	}

	if !user.Live && live {
		s.ChannelMessageSend(discordChannelId, fmt.Sprintf("%s\nhttps://twitcasting.tv/%s", title, user.ScreenId))
		db.Update("TwitcastingUser", userId, "Live", 1)
	} else if user.Live && !live {
		db.Update("TwitcastingUser", userId, "Live", 0)
	}
}

func TwitcastingNotify(name string) {
	defer func() {
		if r := recover(); r != nil {
			tools.DiscordNotify(s, "Twitcasting", name)
			tools.ErrorRecord(r)
		}
	}()

	defer fmt.Printf("%-10s %-20s notification end!\n", name, "Twitcasting")

	userId, discordChannelId := tools.UserData[name]["Twitcasting"]["Id"], tools.UserData[name]["Twitcasting"]["DiscordChannelId"]

	userData := db.FindTwitcastingUser(userId)
	user, err := twitcasting.GetUser(userId)
	if err != nil {
		panic(err)
	}

	baseEmbed := discord.BaseEmbed("Twitcasting", user.Title, user.Url, user.Icon)

	if userData.Title != user.Title {
		baseEmbed.New("", "", "用戶名稱更新了！", "").Change(userData.Title, user.Title).Send(s, discordChannelId)
		db.Update("TwitcastingUser", userId, "Title", user.Title)
	}

	if userData.Description != user.Description {
		baseEmbed.New("", "", "用戶介紹更新了！", "").Change(userData.Description, user.Description).Send(s, discordChannelId)
		db.Update("TwitcastingUser", userId, "Description", user.Description)
	}

	if check, image, err := tools.ImageCheck(userData.Icon, user.Icon); err == nil && check == 0 {
		err = tools.ImageDownload(user.Icon, name, "Twitcasting", "Icon")
		if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "用戶頭貼更新了！", image).Send(s, discordChannelId)
	} else if err != nil {
		panic(err)
	}
}

func TiktokNotify(name string) {
	defer func() {
		if r := recover(); r != nil {
			tools.DiscordNotify(s, "Tiktok", name)
			tools.ErrorRecord(r)
		}
	}()

	defer fmt.Printf("%-10s %-20s notification end!\n", name, "Tiktok")

	userUniqueId, discordChannelId := tools.UserData[name]["Tiktok"]["Id"], testChannelId //tools.UserData[name]["Tiktok"]["DiscordChannelId"]

	user, err := tiktok.GetUser(userUniqueId)
	if errors.Is(err, tiktok.ErrorNoUserData) {
		fmt.Printf("無法順利取得 %s 的 Tiktok 資料！\n", name)
		return
	} else if err != nil {
		panic(err)
	}

	if !db.Find("TiktokUser", "WHERE Id = ?", user.Id) {
		s.ChannelMessageSend(testChannelId, fmt.Sprintf("%s 的 Tiktok UniqueId 可能有變更！", name))
		return
	}

	userData := db.FindTiktokUser(user.Id)

	baseEmbed := discord.BaseEmbed("Tiktok", user.Title, user.Url, user.Icon)

	if userData.Title != user.Title {
		baseEmbed.New("", "", "用戶名稱更新了！", "").Change(userData.Title, user.Title).Send(s, discordChannelId)
		db.Update("TiktokUser", user.Id, "Title", user.Title)
	}

	if userData.ShortId != user.ShortId {
		baseEmbed.New("", "", "用戶短ID更新了！", "").Change(userData.ShortId, user.ShortId).Send(s, discordChannelId)
		db.Update("TiktokUser", user.Id, "ShortId", user.ShortId)
	}

	if userData.UniqueId != user.UniqueId {
		baseEmbed.New("", "", "用戶唯一用戶名更新了！", "").Change(userData.UniqueId, user.UniqueId).Send(s, discordChannelId)
		db.Update("TiktokUser", user.Id, "UniqueId", user.UniqueId)
	}

	if userData.Description != user.Description {
		baseEmbed.New("", "", "用戶介紹更新了！", "").Change(userData.Description, user.Description).Send(s, discordChannelId)
		db.Update("TiktokUser", user.Id, "Description", user.Description)
	}

	if userData.FollowCount != user.FollowCount {
		baseEmbed.New("", "", "用戶追隨數更新了！", "").Change(strconv.Itoa(userData.FollowCount), strconv.Itoa(user.FollowCount)).Send(s, discordChannelId)
		db.Update("TiktokUser", user.Id, "FollowCount", user.FollowCount)
	}

	if check, image, err := tools.ImageCheck(userData.Icon, user.Icon); err == nil && check == 0 {
		err = tools.ImageDownload(user.Icon, name, "Tiktok", "Icon")
		if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "用戶頭貼更新了！", image).Send(s, discordChannelId)
	} else if err != nil {
		panic(err)
	}
}

func FanboxNotify(name string) {
	defer func() {
		if r := recover(); r != nil {
			tools.DiscordNotify(s, "Fanbox", name)
			tools.ErrorRecord(r)
		}
	}()

	defer fmt.Printf("%-10s %-20s notification end!\n", name, "Fanbox")

	userId, discordChannelId := tools.UserData[name]["Fanbox"]["Id"], tools.UserData[name]["Fanbox"]["DiscordChannelId"]

	userData := db.FindFanboxUser(userId)
	user, err := fanbox.GetUser(userId)
	if err != nil {
		panic(err)
	}

	baseEmbed := discord.BaseEmbed("Fanbox", user.Name, user.Url, user.Icon)

	if userData.CreatorId != user.CreatorId {
		baseEmbed.New("", "", "用戶創作者ID更新了！", "").Change(userData.CreatorId, user.CreatorId).Send(s, discordChannelId)
		db.Update("FanboxUser", userId, "CreatorId", user.CreatorId)
	}

	if userData.Name != user.Name {
		baseEmbed.New("", "", "用戶名稱更新了！", "").Change(userData.Name, user.Name).Send(s, discordChannelId)
		db.Update("FanboxUser", userId, "Name", user.Name)
	}

	if userData.Description != user.Description {
		baseEmbed.New("", "", "用戶介紹更新了！", "").Change(userData.Description, user.Description).Send(s, discordChannelId)
		db.Update("FanboxUser", userId, "Description", user.Description)
	}

	if userData.Category != user.Category {
		baseEmbed.New("", "", "用戶分類更新了！", "").Change(userData.Category, user.Category).Send(s, discordChannelId)
		db.Update("FanboxUser", userId, "Category", user.Category)
	}

	if check, image, err := tools.ImageCheck(userData.Icon, user.Icon); err == nil && check == 0 {
		err = tools.ImageDownload(user.Icon, name, "Fanbox", "Icon")
		if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "用戶頭貼更新了！", image).Send(s, discordChannelId)
	} else if err != nil {
		panic(err)
	}

	if check, image, err := tools.ImageCheck(userData.Banner, user.Banner); err == nil && check == 0 {
		err = tools.ImageDownload(user.Banner, name, "Fanbox", "Banner")
		if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "用戶橫幅更新了！", image).Send(s, discordChannelId)
	} else if err != nil {
		panic(err)
	}

	for _, link := range userData.Links {
		if tools.IsContain(user.Links, link) {
			continue
		}

		title, err := tools.GetTitle(link)
		if err != nil {
			panic(err)
		}

		baseEmbed.New(title, link, "刪除了用戶社群連結！", "").Send(s, discordChannelId)
		db.Delete("FanboxLink", "WHERE UserId = ? AND Link = ?", userId, link)
	}

	for _, link := range user.Links {
		if tools.IsContain(userData.Links, link) {
			continue
		}

		title, err := tools.GetTitle(link)
		if err != nil && title == "" {
			panic(err)
		}

		baseEmbed.New(title, link, "新增了用戶社群連結！", "").Send(s, discordChannelId)
		db.Insert("FanboxLink", map[string]any{"UserId": userId, "Link": link})
	}

	for _, items := range fanbox.GroupItem(userData.Items, user.Items) {
		if items.New == nil {
			if items.Old.Type == "image" {
				baseEmbed.New("", "", "刪除了用戶社群相片！", items.Old.Media).Send(s, discordChannelId)
			} else {
				title, err := tools.GetTitle(items.Old.Media)
				if err != nil {
					panic(err)
				}

				baseEmbed.New(title, items.Old.Media, "刪除了用戶社群影片！", "").Send(s, discordChannelId)
			}

			db.Delete("FanboxItem", "WHERE Id = ?", items.Old.Id)
		} else if items.Old == nil {
			if items.New.Type == "image" {
				baseEmbed.New("", "", "新增了用戶社群相片！", items.New.Media).Send(s, discordChannelId)
			} else {
				title, err := tools.GetTitle(items.New.Media)
				if err != nil {
					panic(err)
				}

				baseEmbed.New(title, items.New.Media, "新增了用戶社群影片！", "").Send(s, discordChannelId)
			}

			db.Insert("FanboxItem", items.New.Map())
		}
	}

	oldPlans := db.FindFanboxPlans(userId)
	newPlans, err := fanbox.GetPlan(userId)
	if err != nil {
		panic(err)
	}

	for _, plans := range fanbox.GroupPlan(oldPlans, newPlans) {
		if plans.New == nil {
			image, err := tools.ImageUpload(plans.Old.Image)
			if err != nil {
				panic(err)
			}

			baseEmbed.New(plans.Old.Title, "", "支援方案已被刪除！", image).Send(s, discordChannelId)
			db.Delete("FanboxPlan", "WHERE Id = ?", plans.Old.Id)
			tools.ImageRemove(plans.Old.Image)
		} else if plans.Old == nil {
			err = tools.ImageDownload(plans.New.Image, name, "Fanbox", "Plan", plans.New.Id)
			if err != nil {
				panic(err)
			}

			baseEmbed.New(plans.New.Title, "", "新增了新的支援方案！", plans.New.Image).Send(s, discordChannelId)
			db.Insert("FanboxPlan", plans.New.Map())
		} else {
			if plans.Old.Title != plans.New.Title {
				baseEmbed.New(plans.New.Title, "", "支援方案名稱更新了！", plans.New.Image).Change(plans.Old.Title, plans.New.Title).Send(s, discordChannelId)
				db.Update("FanboxPlan", plans.New.Id, "Title", plans.New.Title)
			}

			if plans.Old.Fee != plans.New.Fee {
				baseEmbed.New(plans.New.Title, "", "支援方案費用更新了！", plans.New.Image).Change(strconv.Itoa(plans.Old.Fee), strconv.Itoa(plans.New.Fee)).Send(s, discordChannelId)
				db.Update("FanboxPlan", plans.New.Id, "Fee", plans.New.Fee)
			}

			if plans.Old.Description != plans.New.Description {
				baseEmbed.New(plans.New.Title, "", "支援方案介紹更新了！", plans.New.Image).Change(plans.Old.Description, plans.New.Description).Send(s, discordChannelId)
				db.Update("FanboxPlan", plans.New.Id, "Description", plans.New.Description)
			}

			if check, image, err := tools.ImageCheck(plans.Old.Image, plans.New.Image); err == nil && check == 0 {
				err = tools.ImageDownload(plans.New.Image, name, "Fanbox", "Plan", plans.New.Id)
				if err != nil {
					panic(err)
				}

				baseEmbed.New(plans.New.Title, "", "支援方案圖片更新了！", image).Send(s, discordChannelId)
			} else if err != nil {
				panic(err)
			}
		}
	}

	oldPosts := db.FindFanboxPosts(userId)
	newPosts, err := fanbox.GetPost(userId)
	if err != nil {
		panic(err)
	}

	for _, posts := range fanbox.GroupPost(oldPosts, newPosts) {
		if posts.New == nil {
			image, err := tools.ImageUpload(posts.Old.Image)
			if err != nil {
				panic(err)
			}

			baseEmbed.New(posts.Old.Title, "", "投稿文章已被刪除！", image).Send(s, discordChannelId)
		} else if posts.Old == nil {
			err = tools.ImageDownload(posts.New.Image, name, "Fanbox", "Post", posts.New.Id)
			if err != nil {
				panic(err)
			}

			baseEmbed.New(posts.New.Title, posts.New.Url, "投稿了新的文章！", posts.New.Image).Send(s, discordChannelId)
			db.Insert("FanboxPost", posts.New.Map())
		} else {
			if posts.Old.UpdatedTime.String() != posts.New.UpdatedTime.String() {
				baseEmbed.New(posts.New.Title, posts.New.Url, "投稿文章更新了！", posts.New.Image).Send(s, testChannelId)
				db.Update("FanboxPost", posts.New.Id, "UpdatedTime", posts.New.UpdatedTime.String())
			}

			if posts.Old.Title != posts.New.Title {
				baseEmbed.New(posts.New.Title, posts.New.Url, "投稿文章名稱更新了！", posts.New.Image).Change(posts.Old.Title, posts.New.Title).Send(s, discordChannelId)
				db.Update("FanboxPost", posts.New.Id, "Title", posts.New.Title)
			}

			if posts.Old.Fee != posts.New.Fee {
				baseEmbed.New(posts.New.Title, posts.New.Url, "投稿文章費用更新了！", posts.New.Image).Change(strconv.Itoa(posts.Old.Fee), strconv.Itoa(posts.New.Fee)).Send(s, discordChannelId)
				db.Update("FanboxPost", posts.New.Id, "Fee", posts.New.Fee)
			}

			// if check, image, err := tools.ImageCheck(posts.Old.Image, posts.New.Image); err == nil && check == 0 {
			// 	err = tools.ImageDownload(posts.New.Image, name, "Fanbox", "Post", posts.New.Id)
			// 	if err != nil {
			// 		panic(err)
			// 	}

			// 	baseEmbed.New(posts.New.Title, posts.New.Url, "投稿文章圖片更新了！", image).Send(s, discordChannelId)
			// } else if err != nil {
			// 	panic(err)
			// }
		}
	}
}

func initial() {
	logFile, err := os.OpenFile("/bot/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}

	log.SetOutput(logFile)

	db, err = sql.ConnectToMySQL(os.Getenv("SQL_USERNAME"), os.Getenv("SQL_PASSWORD"), os.Getenv("SQL_HOST"), os.Getenv("SQL_PORT"), os.Getenv("SQL_DATABASE_NAME"))
	if err != nil {
		log.Fatalf("Error connecting to MySQL: %v", err)
	}

	s, err = discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if handler, ok := commandsHandlers[i.ApplicationCommandData().Name]; ok {
				handler(s, i)
			}
		case discordgo.InteractionMessageComponent:
			parts := strings.Split(i.MessageComponentData().CustomID, ":")

			if parts[0] == "collab" {
				componentsHandlers[parts[0]](s, i, parts[1], parts[2])
			} else {
				componentsHandlers[parts[0]](s, i, "", "")
			}
		}
	})

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Bot is running...")
	})

	for _, command := range commands {
		_, err := s.ApplicationCommandCreate(os.Getenv("DISCORD_APP_ID"), os.Getenv("DISCORD_GUILD_ID"), command)
		if err != nil {
			log.Fatalf("Cannot create slash command: %v", err)
		}
	}

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()
}

func runGo(f func(string), names map[string]int) {
	i := 0

	for name, interval := range names {
		go func(name string, delay int) {
			time.Sleep(time.Duration(delay) * time.Second)
			f(name)

			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				f(name)
			}
		}(name, i*10)
		i++
	}
}

func getChat(names ...string) {
	for _, name := range names {
		for _, videoId := range db.Distinct("livestream", tools.UserData[name]["Youtube"]["Id"]) {
			if videoId == "G22WPiRfTws" {
				continue
			}

			go LiveChat(videoId, tools.UserData[name]["Youtube"]["DiscordChannelId"])
		}
	}
}
