package twitch

import (
	"GoBot/tools"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func GetHoloSchedule(name string) ([]Schedule, error) {
	var streams []Schedule

	url := "https://schedule.hololive.tv/lives/hololive"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []Schedule{}, err
	}

	cookie := http.Cookie{Name: "timezone", Value: "Asia/Tokyo"}
	req.AddCookie(&cookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []Schedule{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return []Schedule{}, fmt.Errorf("HTTP request failed with status code: %d\n%s", resp.StatusCode, string(body))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return []Schedule{}, err
	}

	var keyword string

	if name == "Aqua" {
		keyword = "湊あくあ"
	} else if name == "Shion" {
		keyword = "紫咲シオン"
	}

	var date string

	doc.Find(".col-6.col-sm-4.col-md-3>a, .holodule.navbar-text").Each(func(i int, s *goquery.Selection) {
		if s.HasClass("holodule") {
			date = tools.Regexp(s.Text(), `(\d{2}/\d{2})`, 1)[0][1]
		} else {
			href, _ := s.Attr("href")

			if strings.Contains(href, "twitch") && strings.Contains(s.Text(), keyword) {
				streams = append(streams, Schedule{
					UserId:        name,
					ScheduledTime: getScheduledTime(date, s.Text()),
				})
			}
		}
	})

	return streams, nil
}

func getScheduledTime(d, t string) tools.Time {
	location, _ := time.LoadLocation("Asia/Tokyo")
	year := time.Now().UTC().In(location).Year()
	scheduledTime, _ := time.ParseInLocation("2006/01/02 15:04", fmt.Sprintf("%d/%s %s", year, d, tools.Regexp(t, `(\d{2}:\d{2})`, 1)[0][1]), location)

	return tools.Time(scheduledTime.UTC())
}
