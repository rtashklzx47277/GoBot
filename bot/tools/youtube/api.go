package youtube

import (
	"GoBot/tools"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

var apiKeyList = map[int]string{
	1: os.Getenv("YOUTUBE_API_KEY_1"),
	2: os.Getenv("YOUTUBE_API_KEY_2"),
	3: os.Getenv("YOUTUBE_API_KEY_3"),
	4: os.Getenv("YOUTUBE_API_KEY_4"),
	5: os.Getenv("YOUTUBE_API_KEY_5"),
	6: os.Getenv("YOUTUBE_API_KEY_6"),
	7: os.Getenv("YOUTUBE_API_KEY_7"),
	8: os.Getenv("YOUTUBE_API_KEY_8"),
}

func getData(path string) (*tools.Json, error) {
	// url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/%s&key=%s", path, "AIzaSyDyLkKGgzBYiEjORAkXNmnNnlwotAD-RHI")
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/%s&key=%s", path, apiKeyList[time.Now().Hour()/3+1])
	reader, err := tools.Get(url).Do()
	if err != nil {
		return &tools.Json{}, err
	}

	data, err := tools.ToJson(reader)
	if err != nil {
		return &tools.Json{}, err
	}

	return data, nil
}

func GetChannel(channelId string) (Channel, error) {
	path := fmt.Sprintf("channels?part=brandingSettings,snippet,statistics&id=%s", channelId)
	data, err := getData(path)
	if err != nil {
		return Channel{}, err
	}

	if !data.Exist("items") {
		return Channel{}, errors.New("fail to get data")
	}

	item := data.Get("items").Index(0)
	channel := Channel{
		Id:              item.Get("id").String(),
		CustomId:        item.Get("snippet").Get("customUrl").Slice(1, -1),
		Url:             fmt.Sprintf("https://www.youtube.com/channel/%s", item.Get("id").String()),
		Title:           item.Get("snippet").Get("title").String(),
		Description:     item.Get("snippet").Get("description").String(),
		Icon:            item.Get("snippet").Get("thumbnails").Image(),
		Banner:          item.Get("brandingSettings").Get("image").Get("bannerExternalUrl").String(),
		SubscriberCount: item.Get("statistics").Get("subscriberCount").Int(),
		ViewCount:       item.Get("statistics").Get("viewCount").Int(),
	}

	return channel, nil
}

func GetChannelSections(channelId string) ([]map[string]any, error) {
	path := fmt.Sprintf("channelSections?part=contentDetails,snippet&channelId=%s", channelId)
	data, err := getData(path)
	if err != nil {
		return []map[string]any{}, err
	}

	var sections []map[string]any

	for _, item := range data.Get("items").JsonArray() {
		var section map[string]any
		sectionType := item.Get("snippet").Get("type").String()

		if sectionType == "channelsectiontypeundefined" {
			continue
		} else if sectionType == "multiplechannels" || sectionType == "singleplaylist" || sectionType == "multipleplaylists" {
			var sectionContent []any

			if sectionType == "multiplechannels" {
				sectionContent = item.Get("contentDetails").Get("channels").Array()
			} else {
				sectionContent = item.Get("contentDetails").Get("playlists").Array()
			}

			section = map[string]any{
				"Type":    sectionType,
				"Content": sectionContent,
			}
		} else {
			section = map[string]any{
				"Type": sectionType,
			}
		}

		sections = append(sections, section)
	}

	return sections, nil
}

func GetPlaylists(channelId string) ([]Playlist, error) {
	var playlists []Playlist
	var pageToken string

	for {
		path := fmt.Sprintf("playlists?part=snippet,status&channelId=%s&maxResults=50&pageToken=%s", channelId, pageToken)
		data, err := getData(path)
		if err != nil {
			return []Playlist{}, err
		}

		if !data.Exist("items") {
			return []Playlist{}, errors.New("fail to get data")
		}

		pageToken = data.Get("nextPageToken").String()

		for _, item := range data.Get("items").JsonArray() {
			playlist := Playlist{
				Id:          item.Get("id").String(),
				Url:         fmt.Sprintf("https://www.youtube.com/playlist?list=%s", item.Get("id").String()),
				Title:       item.Get("snippet").Get("title").String(),
				Description: item.Get("snippet").Get("description").String(),
				Thumbnail:   item.Get("snippet").Get("thumbnails").Image(),
			}
			playlist.Author.Id = channelId
			playlists = append(playlists, playlist)
		}

		if pageToken == "" {
			break
		}
	}

	return playlists, nil
}

func GetPlaylistItems(playlistId string, num int) ([]Video, error) {
	var videos []Video
	var pageToken string

	for {
		path := fmt.Sprintf("playlistItems?part=snippet&playlistId=%s&maxResults=%d&pageToken=%s", playlistId, num, pageToken)
		data, err := getData(path)
		if err != nil {
			if strings.HasPrefix(playlistId, "UUMO") {
				return []Video{}, nil
			}

			return []Video{}, err
		}

		if !data.Exist("items") {
			return []Video{}, errors.New("fail to get data")
		}

		pageToken = data.Get("nextPageToken").String()

		for _, item := range data.Get("items").JsonArray() {
			video := Video{
				Id:        item.Get("snippet").Get("resourceId").Get("videoId").String(),
				Url:       fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.Get("snippet").Get("resourceId").Get("videoId").String()),
				Title:     item.Get("snippet").Get("title").String(),
				Thumbnail: item.Get("snippet").Get("thumbnails").Image(),
			}

			videos = append(videos, video)
		}

		if pageToken == "" || num != 50 {
			break
		}
	}

	return videos, nil
}

func GetVideo(videoId string) (Video, error) {
	path := fmt.Sprintf("videos?part=contentDetails,liveStreamingDetails,snippet,statistics&id=%s", videoId)
	data, err := getData(path)
	if err != nil {
		return Video{}, err
	}

	if !data.Exist("items") {
		return Video{}, errors.New("fail to get data")
	}

	var item *tools.Json

	if len(data.Get("items").JsonArray()) != 0 {
		item = data.Get("items").Index(0)
	} else {
		item = nil
	}

	return getVideoStruct(item, videoId), nil
}

func GetVideos(videoIds []string) ([]Video, error) {
	var videos []Video
	length := len(videoIds)

	for n := 0; n < length; n += 50 {
		end := n + 50

		if end > length {
			end = length
		}

		path := fmt.Sprintf("videos?part=contentDetails,liveStreamingDetails,snippet,statistics&id=%s", strings.Join(videoIds[n:end], ","))
		data, err := getData(path)
		if err != nil {
			return []Video{}, err
		}

		if !data.Exist("items") {
			return []Video{}, errors.New("fail to get data")
		}

		dataList := data.Get("items").JsonArray()

		for _, videoId := range videoIds[n:end] {
			var item *tools.Json

			if len(dataList) != 0 && videoId == dataList[0].Get("id").String() {
				item = dataList[0]
				dataList = dataList[1:]
			} else {
				item = nil
			}

			videos = append(videos, getVideoStruct(item, videoId))
		}
	}

	return videos, nil
}

func GetComments(target, Id string) ([]Comment, error) {
	var comments []Comment
	var request, filter, pageToken string

	if target == "channel" {
		request, filter = "commentThreads", "allThreadsRelatedToChannelId"
	} else if target == "video" {
		request, filter = "commentThreads", "videoId"
	} else if target == "reply" {
		request, filter = "comments", "parentId"
	}

	for {
		path := fmt.Sprintf("%s?part=snippet&%s=%s&maxResults=100&textFormat=plainText&pageToken=%s", request, filter, Id, pageToken)
		data, err := getData(path)
		if err != nil {
			return []Comment{}, err
		}

		if !data.Exist("items") {
			return []Comment{}, errors.New("fail to get data")
		}

		pageToken = data.Get("nextPageToken").String()
		dataList := data.Get("items").JsonArray()

		if len(dataList) == 0 {
			pageToken = ""
		}

		for _, item := range dataList {
			var snippet *tools.Json

			if target == "channel" || target == "video" {
				snippet = item.Get("snippet").Get("topLevelComment").Get("snippet")
			} else if target == "reply" {
				snippet = item.Get("snippet")
			}

			commentTime := snippet.Get("publishedAt").Time()
			authorId := snippet.Get("authorChannelId").Get("value").String()

			if !commentTime.InRange(1) {
				pageToken = ""
				break
			} else if _, ok := tools.ChannelList[authorId]; ok {
				var id string

				if target == "channel" || target == "video" {
					id = item.Get("id").String()
				} else if target == "reply" {
					id = item.Get("id").Split(".")[1]
				}

				comment := Comment{
					Id:            id,
					Text:          snippet.Get("textDisplay").String(),
					PublishedTime: commentTime,
					UpdatedTime:   snippet.Get("updatedAt").Time(),
				}

				comment.Author.Id = authorId
				comment.Author.Url = fmt.Sprintf("https://www.youtube.com/channel/%s", authorId)

				if target == "channel" || target == "video" {
					comment.Video.Id = snippet.Get("videoId").String()
					comment.Video.Url = fmt.Sprintf("https://www.youtube.com/watch?v=%s", comment.Video.Id)
				} else if target == "reply" {
					comment.ParentId = Id
				}

				comments = append(comments, comment)
			}
		}

		if pageToken == "" {
			break
		}
	}

	return comments, nil
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
