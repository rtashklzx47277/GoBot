package youtube

import (
	"GoBot/tools"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetCollab(channelId string) ([]string, error) {
	var videoIdList []string

	url := fmt.Sprintf("https://holodex.net/api/v2/channels/%s/collabs?type=stream,placeholder&include=clips,live_info&limit=3", channelId)
	reader, err := tools.Get(url).AddHeader("Accept", "application/json, text/plain, */*").
		AddHeader("Accept-Encoding", "deflate").
		AddHeader("Referer", fmt.Sprintf("https://holodex.net/channel/%s/collabs", channelId)).
		AddHeader("User-Agent", tools.UserAgent).Do()
	if err != nil {
		return []string{}, err
	}

	data, err := tools.ToJson(reader)
	if err != nil {
		return []string{}, err
	}

	for _, video := range data.JsonArray() {
		videoIdList = append(videoIdList, video.Get("id").String())
	}

	return videoIdList, nil
}

func GetSchedule(name string) ([]Video, error) {
	var videoIdList []string
	var streams []Video

	reader, err := tools.Get("https://schedule.hololive.tv/lives/hololive").AddCookie("timezone", "Asia/Tokyo").Do()
	if err != nil {
		return []Video{}, err
	}

	doc, err := tools.ToDocument(reader)
	if err != nil {
		return []Video{}, err
	}

	doc.Find(".col-6.col-sm-4.col-md-3>a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")

		if strings.Contains(href, "youtube") {
			if videoId, ok := tools.Regexp(href, `\?v=([\w\-_]{11})`); ok {
				videoIdList = append(videoIdList, videoId)
			}
		}
	})

	videos, err := GetVideos(videoIdList)
	if err != nil {
		return []Video{}, err
	}

	for _, video := range videos {
		if checkCollab(name, video) {
			streams = append(streams, video)
		}
	}

	return streams, nil
}

func checkCollab(name string, video Video) bool {
	var channelId string
	var keywords []string

	if name == "Aqua" {
		channelId = "UC1opHUrw8rvnsadT-iGp7Cg"
		keywords = []string{"湊あくあ", "Aqua"}
	} else if name == "Shion" {
		channelId = "UCXTpFs_3PqI41qX2d9tL2Rw"
		keywords = []string{"紫咲シオン", "Shion"}
	}

	if video.Author.Id == channelId {
		return false
	}

	for _, keyword := range keywords {
		if strings.Contains(video.Title, keyword) || strings.Contains(video.Description, keyword) {
			return true
		}
	}

	return false
}
