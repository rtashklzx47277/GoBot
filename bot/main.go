package main

import (
	"GoBot/tools"
	"GoBot/tools/discord"
	"GoBot/tools/fanbox"
	"GoBot/tools/sql"
	"GoBot/tools/twitcasting"
	"GoBot/tools/twitch"
	"GoBot/tools/youtube"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func main() {
	initial()

	runGo(YoutubeStreamNotify, 1, "Aqua", "Shion")
	runGo(YoutubeNotify, 10, "Aqua", "Shion")
	runGo(Collab, 10, "Aqua", "Shion")
	runGo(TwitchStreamNotify, 1, "Aqua", "Shion")
	runGo(TwitchNotify, 10, "Aqua", "Shion")
	runGo(News, 10, "Aqua", "Shion")

	select {}

	// LiveChatbyOriginal("eyIubrMru0s")
}

func YoutubeStreamNotify(name string) {
	defer func() {
		if r := recover(); r != nil {
			tools.DiscordNotify(s, "Youtube Live", name)
			tools.ErrorRecord(r)
		}
	}()

	defer fmt.Printf("%-10s %-20s notification end!\n", name, "Youtube Live")

	channelId, discordChannelId := userDataMap[name]["Youtube"]["Id"], userDataMap[name]["Youtube"]["DiscordChannelId"]

	channel, err := youtube.GetChannel(channelId)
	if err != nil {
		panic(err)
	}

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

		if isContain(videoIds, videoId) {
			continue
		}

		video, err := youtube.GetVideo(videoId)
		if err != nil {
			panic(err)
		}

		err = tools.ImageDownload(video.Thumbnail, "Youtube", channelId, "Video", video.Id)
		if err != nil {
			panic(err)
		}

		status := ""

		if i >= index {
			video.Member = true
			status = "member"
		}

		// if video.Live {
		// 	go LiveChat(video, channel)
		// }

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

				baseEmbed.New(old.Title, old.Url, "預定直播已被取消了！", thumbnail).Send(s, discordChannelId)
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

	channelId, discordChannelId := userDataMap[name]["Youtube"]["Id"], userDataMap[name]["Youtube"]["DiscordChannelId"]

	channelData := db.FindChannel(channelId)
	channel, err := youtube.GetChannel(channelId)
	if err != nil {
		panic(err)
	}

	baseEmbed := discord.BaseEmbed("Youtube", channel.Title, channel.Url, channel.Icon)

	if channelData.SubscriberCount/10000 < channel.SubscriberCount/10000 {
		baseEmbed.New("", "", fmt.Sprintf("Youtube訂閱者數已突破%d萬人了！", channel.SubscriberCount/10000), channel.Icon).Send(s, discordChannelId)
		db.Update("Channel", channelId, "SubscriberCount", channel.SubscriberCount)
	}

	if channelData.ViewCount/50000000 < channel.ViewCount/50000000 {
		baseEmbed.New("", "", fmt.Sprintf("Youtube觀看次數已突破%s億次了！", fmt.Sprintf("%.1f", float64(channel.ViewCount)/100000000)), channel.Icon).Send(s, discordChannelId)
		db.Update("Channel", channelId, "ViewCount", channel.ViewCount)
	}

	if channelData.CustomId != channel.CustomId {
		baseEmbed.New("", "", "頻道ID更新了！", "").Change(channelData.CustomId, channel.CustomId).Send(s, discordChannelId)
		db.Update("Channel", channelId, "CustomId", channel.CustomId)
	}

	if channelData.Title != channel.Title {
		baseEmbed.New("", "", "頻道名稱更新了！", "").Change(channelData.Title, channel.Title).Send(s, discordChannelId)
		db.Update("Channel", channelId, "Title", channel.Title)
	}

	if channelData.Description != channel.Description {
		baseEmbed.New("", "", "頻道介紹更新了！", "").Change(channelData.Description, channel.Description).Send(s, discordChannelId)
		db.Update("Channel", channelId, "Description", channel.Description)
	}

	if check, image, err := tools.ImageCheck(channelData.Icon, channel.Icon); err == nil && check == 0 {
		err = tools.ImageDownload(channel.Icon, "Youtube", channelId, "Icon", channelId)
		if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "頻道頭貼更新了！", image).Send(s, discordChannelId)
	} else if err != nil {
		panic(err)
	}

	if check, image, err := tools.ImageCheck(channelData.Banner, channel.Banner); err == nil && check == 0 {
		err = tools.ImageDownload(channel.Banner, "Youtube", channelId, "Banner", channelId)
		if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "頻道橫幅更新了！", image).Send(s, discordChannelId)
	} else if err != nil {
		panic(err)
	}

	oldMusic := db.FindMusic(channelId)
	newMusic, err := youtube.GetVideos(db.Distinct("music", channelId))
	if err != nil {
		panic(err)
	}

	for i := range oldMusic {
		old, new := oldMusic[i], newMusic[i]

		if (new.ViewCount < 1000000 && new.ViewCount/100000 > old.ViewCount/100000) || (new.ViewCount >= 1000000 && new.ViewCount/500000 > old.ViewCount/500000) {
			baseEmbed.New(new.Title, new.Url, fmt.Sprintf("影片觀看次數已突破%d萬次了！", new.ViewCount/10000), new.Thumbnail).Send(s, discordChannelId)
			db.Update("Video", new.Id, "ViewCount", new.ViewCount)
		}
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
			err = tools.ImageDownload(new.Thumbnail, "Youtube", channelId, "Video", new.Id)
			if err != nil {
				panic(err)
			}

			commentIds := db.Distinct("comment", channelId)
			comments, err := youtube.GetComments("video", new.Id)
			if err != nil {
				panic(err)
			}

			for _, comment := range comments {
				if !isContain(commentIds, comment.Id) {
					db.Insert("Comment", comment.Map())
				}
			}
			// turn to public

			baseEmbed.New(new.Title, new.Url, "影片已設為公開了！", new.Thumbnail).Send(s, discordChannelId)
			db.Update("Video", new.Id, "Title", new.Title, "Description", new.Description, "Length", new.Length.String(), "ViewCount", new.ViewCount, "LiveStatus", new.LiveStatus, "PublishedTime", new.PublishedTime.String(), "ScheduledTime", new.ScheduledTime.String(), "StartTime", new.StartTime.String(), "EndTime", new.EndTime.String(), "Comment", new.Comment, "Live", new.Live, "Private", new.Private)
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
				baseEmbed.New(new.Title, new.Url, "影片長度更新了！", new.Thumbnail).Change(old.Length.String(), new.Length.String()).Send(s, testChannelId)
				db.Update("Video", new.Id, "Length", new.Length.String())
			}

			if old.PublishedTime != new.PublishedTime {
				db.Update("Video", new.Id, "PublishedTime", new.PublishedTime.String())
			}

			if old.ViewCount/100000 != new.ViewCount/100000 && !old.Music {
				baseEmbed.New(new.Title, new.Url, fmt.Sprintf("影片觀看次數已突破%d萬次了！", new.ViewCount/10000), new.Thumbnail).Send(s, testChannelId)
				db.Update("Video", new.Id, "ViewCount", new.ViewCount)
			}

			if new.LiveStatus != 0 {
				check, image, err := tools.ImageCheck(old.Thumbnail, new.Thumbnail)
				if err != nil {
					panic(err)
				}

				if check != 1 {
					err = tools.ImageDownload(new.Thumbnail, "Youtube", channelId, "Video", new.Id)
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
		if isContain(videoIds, video.Id) {
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
		if isContain(videoIds, video.Id) {
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
			err = tools.ImageDownload(playlists.New.Thumbnail, "Youtube", channelId, "Playlist", playlists.New.Id)
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
				err = tools.ImageDownload(playlists.New.Thumbnail, "Youtube", channelId, "Playlist", playlists.New.Id)
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

						db.Insert("Video", video.Map())
					}

					baseEmbed.New(playlistItems.New.Title, playlistItems.New.Url, fmt.Sprintf("追加新影片至「%s」中！", playlists.New.Title), playlistItems.New.Thumbnail).Send(s, discordChannelId)
					db.Insert("PlaylistItem", map[string]any{"PlaylistId": playlists.New.Id, "VideoId": playlistItems.New.Id})
				}
			}
		}
	}

	commentIds := db.Distinct("comment", channelId)
	replyIds := db.Distinct("comment", channelId)
	comments, err := youtube.GetComments("channel", channelId)
	if err != nil {
		panic(err)
	}

	for _, comment := range comments {
		if !isContain(commentIds, comment.Id) {
			comment = db.CompelteComment(comment)
			s.ChannelMessageSend(testChannelId, fmt.Sprintf("「[%s](<%s>)」在「[%s](<%s>)」的影片「[%s](<%s>)」中發表留言：\n> %s",
				comment.Author.Title, comment.Author.Url, channel.Title, channel.Url, comment.Video.Title, comment.Video.Url, strings.Replace(comment.Text, "\n", "\n> ", -1)))
			db.Insert("Comment", comment.Map())
		}

		replies, err := youtube.GetComments("reply", comment.Id)
		if err != nil {
			panic(err)
		}

		for _, reply := range replies {
			if !isContain(replyIds, reply.Id) {
				reply = db.CompelteComment(reply)
				s.ChannelMessageSend(testChannelId, fmt.Sprintf("「[%s](<%s>)」在「[%s](<%s>)」的影片「[%s](<%s>)」中發表留言：\n> %s",
					reply.Author.Title, reply.Author.Url, channel.Title, channel.Url, reply.Video.Title, reply.Video.Url, strings.Replace(reply.Text, "\n", "\n> ", -1)))
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
		if !isContain(postIds, post.Id) {
			var description string

			if post.Member {
				description = "會員限定"
			}

			if post.Renderer.Type == "Image" {
				for i, image := range post.Renderer.Images {
					err = tools.ImageDownload(image, "Youtube", channelId, "Post", fmt.Sprintf("%s_%d", post.Id, i+1))
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
}

func Collab(name string) {
	defer func() {
		if r := recover(); r != nil {
			tools.DiscordNotify(s, "Collab", name)
			tools.ErrorRecord(r)
		}
	}()

	defer fmt.Printf("%-10s %-20s notification end!\n", name, "Collab")

	channelId, discordChannelId := userDataMap[name]["Youtube"]["Id"], userDataMap[name]["Youtube"]["DiscordChannelId"]

	oldIds := db.Distinct("collab", channelId)
	newIds, err := youtube.GetCollab(channelId)
	if err != nil {
		panic(err)
	}

	for _, videoId := range newIds {
		if isContain(oldIds, videoId) {
			continue
		}

		video, err := youtube.GetVideo(videoId)
		if err != nil {
			panic(err)
		}

		err = tools.ImageDownload(video.Thumbnail, "Youtube", channelId, "Collab", video.Id)
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
		if isContain(collabIds, video.Id) || isContain(oldIds, video.Id) {
			continue
		}

		s.ChannelMessageSendComplex(testChannelId, &discordgo.MessageSend{
			Embed:      (*discordgo.MessageEmbed)(discord.BaseEmbed("Youtube", "", "", "").NewNotify("collab", video)),
			Components: getComponent(name, video.Id),
		})

		collabIds = append(collabIds, video.Id)
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

	userId, discordChannelId := userDataMap[name]["Twitch"]["Id"], userDataMap[name]["Twitch"]["DiscordChannelId"]

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

	userId, discordChannelId := userDataMap[name]["Twitch"]["Id"], userDataMap[name]["Twitch"]["DiscordChannelId"]

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
		err = tools.ImageDownload(user.Icon, "Twitch", userId, "Icon", userId)
		if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "用戶頻道頭貼更新了！", image).Send(s, discordChannelId)
	} else if err != nil {
		panic(err)
	}

	if check, image, err := tools.ImageCheck(userData.Thumbnail, user.Thumbnail); err == nil && check == 0 {
		err = tools.ImageDownload(user.Thumbnail, "Twitch", userId, "Thumbnail", userId)
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
			err = tools.ImageDownload(badges.New.Image, "Twitch", userId, "Badge", badges.New.Id)
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
			err = tools.ImageDownload(stamps.New.Image, "Twitch", userId, "Stamp", stamps.New.Id)
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

	userId, discordChannelId := userDataMap[name]["Twitcasting"]["Id"], userDataMap[name]["Twitcasting"]["DiscordChannelId"]
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

	userId, discordChannelId := userDataMap[name]["Twitcasting"]["Id"], userDataMap[name]["Twitcasting"]["DiscordChannelId"]

	userData := db.FindTwitcastingUser(userId)
	user, err := twitcasting.GetUser(userId)
	if err != nil {
		panic(err)
	}

	baseEmbed := discord.BaseEmbed("Twitcasting", user.Title, user.Url, user.Icon)

	if userData.Title != user.Title {
		baseEmbed.New("", "", "頻道名稱更新了！", "").Change(userData.Title, user.Title).Send(s, discordChannelId)
		db.Update("TwitcastingUser", userId, "Title", user.Title)
	}

	if userData.Description != user.Description {
		baseEmbed.New("", "", "頻道介紹更新了！", "").Change(userData.Description, user.Description).Send(s, discordChannelId)
		db.Update("TwitcastingUser", userId, "Description", user.Description)
	}

	if check, image, err := tools.ImageCheck(userData.Icon, user.Icon); err == nil && check == 0 {
		err = tools.ImageDownload(user.Icon, "Twitcasting", userId, "Icon", userId)
		if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "頻道頭貼更新了！", image).Send(s, discordChannelId)
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

	userId, discordChannelId := userDataMap[name]["Fanbox"]["Id"], userDataMap[name]["Fanbox"]["DiscordChannelId"]

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

	if check, image, err := tools.ImageCheck(userData.Icon, user.Icon); err == nil && check == 0 {
		err = tools.ImageDownload(user.Icon, "FanboxUser", userId, "Icon", userId)
		if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "用戶頭貼更新了！", image).Send(s, discordChannelId)
	} else if err != nil {
		panic(err)
	}

	if check, image, err := tools.ImageCheck(userData.Banner, user.Banner); err == nil && check == 0 {
		err = tools.ImageDownload(user.Icon, "FanboxUser", userId, "Banner", userId)
		if err != nil {
			panic(err)
		}

		baseEmbed.New("", "", "用戶橫幅更新了！", image).Send(s, discordChannelId)
	} else if err != nil {
		panic(err)
	}
}

func News(name string) {
	defer func() {
		if r := recover(); r != nil {
			tools.DiscordNotify(s, "News", name)
			tools.ErrorRecord(r)
		}
	}()

	defer fmt.Printf("%-10s %-20s notification end!\n", name, "News")

	userId, discordChannelId := userDataMap[name]["News"]["Id"], userDataMap[name]["News"]["DiscordChannelId"]

	newsList, err := tools.GetNews(userId)
	if err != nil {
		panic(err)
	}

	for _, news := range newsList {
		if !db.Find("News", "WHERE Url = ? AND Mention = ?", news.Url, userId) {
			s.ChannelMessageSend(discordChannelId, fmt.Sprintf("%s\n%s", news.Title, news.Url))
			db.Insert("News", news.Map())
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
		fmt.Println("Bot is running.")
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

func runGo(f func(string), interval int, names ...string) {
	for _, name := range names {
		go func(name string) {
			f(name)

			ticker := time.NewTicker(time.Duration(interval) * time.Minute)
			defer ticker.Stop()

			for range ticker.C {
				f(name)
			}
		}(name)
	}
}
