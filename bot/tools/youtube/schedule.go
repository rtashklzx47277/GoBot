package youtube

import (
	"GoBot/tools"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetSchedule(name string) ([]Video, error) {
	var videoIdList []string
	var streams []Video

	url := "https://schedule.hololive.tv/lives/hololive"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []Video{}, err
	}

	cookie := http.Cookie{Name: "timezone", Value: "Asia/Tokyo"}
	req.AddCookie(&cookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []Video{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return []Video{}, fmt.Errorf("HTTP request failed with status code: %d\n%s", resp.StatusCode, string(body))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return []Video{}, err
	}

	doc.Find(".col-6.col-sm-4.col-md-3>a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")

		if strings.Contains(href, "youtube") {
			videoIdList = append(videoIdList, tools.Regexp(href, `(\?v=|live\/)([\w\-_]{11})`, 1)[0][2])
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

func IsExist(list []Video, target Video) bool {
	for _, element := range list {
		if element == target {
			return true
		}
	}

	return false
}
