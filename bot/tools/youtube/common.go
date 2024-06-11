package youtube

import (
	"GoBot/tools"
	"fmt"
)

type Channel struct {
	Id              string
	CustomId        string
	Url             string
	Title           string
	Description     string
	Icon            string
	Banner          string
	SubscriberCount int
	ViewCount       int
}

type Playlist struct {
	Id          string
	Url         string
	Title       string
	Description string
	Thumbnail   string
	Author      Channel
}

type Video struct {
	Id            string
	Url           string
	Title         string
	Description   string
	Thumbnail     string
	Length        tools.Duration
	ViewCount     int
	LiveStatus    int
	PublishedTime tools.Time
	ScheduledTime tools.Time
	StartTime     tools.Time
	EndTime       tools.Time
	Comment       bool
	Member        bool
	Live          bool
	Private       bool
	Music         bool
	Author        Channel
}

type Comment struct {
	Id            string
	ParentId      string
	Text          string
	PublishedTime tools.Time
	UpdatedTime   tools.Time
	Video         Video
	Author        Channel
	Canceled      bool
}

type Post struct {
	Id       string
	Url      string
	Text     string
	Member   bool
	Renderer Renderer
}

type Renderer struct {
	Type     string
	Images   []string
	Video    Video
	Playlist Playlist
	Choices  []Choice
}

type Choice struct {
	Text      string
	Image     string
	isCorrect bool
}

type ZipVideo struct {
	Old *Video
	New *Video
}

type ZipPlaylist struct {
	Old *Playlist
	New *Playlist
}

type ZipComment struct {
	Old *Comment
	New *Comment
}

func (channel Channel) Map() map[string]any {
	channelMap := map[string]any{
		"Id":              channel.Id,
		"CustomId":        channel.CustomId,
		"Title":           channel.Title,
		"Description":     channel.Description,
		"SubscriberCount": channel.SubscriberCount,
		"ViewCount":       channel.ViewCount,
	}

	return channelMap
}

func (playlist Playlist) Map() map[string]any {
	playlistMap := map[string]any{
		"Id":          playlist.Id,
		"ChannelId":   playlist.Author.Id,
		"Title":       playlist.Title,
		"Description": playlist.Description,
	}

	return playlistMap
}

func (video Video) Map() map[string]any {
	var videoMap map[string]any

	if video.Private {
		videoMap = map[string]any{
			"Title":         nil,
			"Description":   nil,
			"Length":        nil,
			"ViewCount":     nil,
			"LiveStatus":    nil,
			"PublishedTime": nil,
			"ScheduledTime": nil,
			"StartTime":     nil,
			"EndTime":       nil,
		}
	} else {
		videoMap = map[string]any{
			"Title":         video.Title,
			"Description":   video.Description,
			"Length":        video.Length.String(),
			"ViewCount":     video.ViewCount,
			"LiveStatus":    video.LiveStatus,
			"PublishedTime": video.PublishedTime.String(),
		}

		if video.Live {
			videoMap["ScheduledTime"] = video.ScheduledTime.String()
		} else {
			videoMap["ScheduledTime"] = nil
		}

		if video.StartTime != (tools.Time{}) {
			videoMap["StartTime"] = video.StartTime.String()
		} else {
			videoMap["StartTime"] = nil
		}

		if video.EndTime != (tools.Time{}) {
			videoMap["EndTime"] = video.EndTime.String()
		} else {
			videoMap["EndTime"] = nil
		}
	}

	videoMap["Id"] = video.Id
	videoMap["ChannelId"] = video.Author.Id
	videoMap["Comment"] = video.Comment
	videoMap["Member"] = video.Member
	videoMap["Live"] = video.Live
	videoMap["Private"] = video.Private
	videoMap["Music"] = video.Music

	return videoMap
}

func (comment Comment) Map() map[string]any {
	commentMap := map[string]any{
		"Id":            comment.Id,
		"Text":          comment.Text,
		"PublishedTime": comment.PublishedTime.String(),
		"UpdatedTime":   comment.UpdatedTime.String(),
		"ChannelId":     comment.Author.Id,
		"VideoId":       comment.Video.Id,
		"Canceled":      comment.Canceled,
	}

	if comment.ParentId != "" {
		commentMap["ParentId"] = comment.ParentId
	}

	return commentMap
}

func getVideoStruct(data *tools.Json, videoId string) Video {
	if data == nil {
		return Video{Id: videoId, Private: true, Author: Channel{Id: ""}}
	}

	video := Video{
		Id:            videoId,
		Url:           fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoId),
		Title:         data.Get("snippet").Get("title").String(),
		Description:   data.Get("snippet").Get("description").String(),
		Thumbnail:     data.Get("snippet").Get("thumbnails").Image(),
		Length:        data.Get("contentDetails").Get("duration").Duration(),
		ViewCount:     data.Get("statistics").Get("viewCount").Int(),
		PublishedTime: data.Get("snippet").Get("publishedAt").Time(),
		ScheduledTime: data.Get("liveStreamingDetails").Get("scheduledStartTime").Time(),
		StartTime:     data.Get("liveStreamingDetails").Get("actualStartTime").Time(),
		EndTime:       data.Get("liveStreamingDetails").Get("actualEndTime").Time(),
		Comment:       data.Get("statistics").Exist("commentCount"),
	}

	switch data.Get("snippet").Get("liveBroadcastContent").String() {
	case "none":
		video.LiveStatus = 0
	case "upcoming":
		video.LiveStatus = 1
	case "live":
		video.LiveStatus = 2
	}

	if video.ScheduledTime != (tools.Time{}) {
		video.Live = true
	}

	video.Author.Id = data.Get("snippet").Get("channelId").String()
	video.Author.Url = fmt.Sprintf("https://www.youtube.com/channel/%s", video.Author.Id)

	return video
}

func GroupVideo(old, new []Video) []ZipVideo {
	result := []ZipVideo{}

	videoMap := map[string]bool{}
	oldMap, newMap := map[string]Video{}, map[string]Video{}

	for _, video := range old {
		videoMap[video.Id] = true
		oldMap[video.Id] = video
	}

	for _, video := range new {
		videoMap[video.Id] = true
		newMap[video.Id] = video
	}

	for videoId := range videoMap {
		oldVideo, ok1 := oldMap[videoId]
		newVideo, ok2 := newMap[videoId]

		if ok1 && ok2 {
			result = append(result, ZipVideo{Old: &oldVideo, New: &newVideo})
		} else if ok1 && !ok2 {
			result = append(result, ZipVideo{Old: &oldVideo, New: nil})
		} else if !ok1 && ok2 {
			result = append(result, ZipVideo{Old: nil, New: &newVideo})
		}
	}

	return result
}

func GroupPlaylist(old, new []Playlist) []ZipPlaylist {
	result := []ZipPlaylist{}

	playlistMap := map[string]bool{}
	oldMap, newMap := map[string]Playlist{}, map[string]Playlist{}

	for _, playlist := range old {
		playlistMap[playlist.Id] = true
		oldMap[playlist.Id] = playlist
	}

	for _, playlist := range new {
		playlistMap[playlist.Id] = true
		newMap[playlist.Id] = playlist
	}

	for playlistId := range playlistMap {
		oldPlaylist, ok1 := oldMap[playlistId]
		newPlaylist, ok2 := newMap[playlistId]

		if ok1 && ok2 {
			result = append(result, ZipPlaylist{Old: &oldPlaylist, New: &newPlaylist})
		} else if ok1 && !ok2 {
			result = append(result, ZipPlaylist{Old: &oldPlaylist, New: nil})
		} else if !ok1 && ok2 {
			result = append(result, ZipPlaylist{Old: nil, New: &newPlaylist})
		}
	}

	return result
}

func GroupComment(old, new []Comment) []ZipComment {
	result := []ZipComment{}

	commentMap := map[string]bool{}
	oldMap, newMap := map[string]Comment{}, map[string]Comment{}

	for _, comment := range old {
		commentMap[comment.Id] = true
		oldMap[comment.Id] = comment
	}

	for _, comment := range new {
		commentMap[comment.Id] = true
		newMap[comment.Id] = comment
	}

	for commentId := range commentMap {
		oldComment, ok1 := oldMap[commentId]
		newComment, ok2 := newMap[commentId]

		if ok1 && ok2 {
			result = append(result, ZipComment{Old: &oldComment, New: &newComment})
		} else if ok1 && !ok2 {
			result = append(result, ZipComment{Old: &oldComment, New: nil})
		} else if !ok1 && ok2 {
			result = append(result, ZipComment{Old: nil, New: &newComment})
		}
	}

	return result
}
