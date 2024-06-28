package tools

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type News struct {
	Title         string
	PublishedTime Time
	Url           string
	Mention       string
}

func GetNews(keyword string) ([]News, error) {
	var newsList []News

	news, err := getPrtimes(keyword)
	if err != nil {
		return []News{}, fmt.Errorf("error occurred while getting Prtimes data: %v", err)
	}
	newsList = append(newsList, news...)

	news, err = getPanora(keyword)
	if err != nil {
		return []News{}, fmt.Errorf("error occurred while getting Panora data: %v", err)
	}
	newsList = append(newsList, news...)

	news, err = getRealsound(keyword)
	if err != nil {
		return []News{}, fmt.Errorf("error occurred while getting Realsound data: %v", err)
	}

	newsList = append(newsList, news...)

	return newsList, nil
}

func getPrtimes(keyword string) ([]News, error) {
	var newsList []News

	doc, err := getDoc(fmt.Sprintf("https://prtimes.jp/main/action.php?page=searchkey&search_word=%s", url.QueryEscape(keyword)))
	if err != nil {
		return []News{}, err
	}

	for _, node := range doc.Find("article").Nodes[:3] {
		nodeDoc := goquery.NewDocumentFromNode(node)

		t, _ := nodeDoc.Find("time").Attr("datetime")
		publishedTime := stringToTime(t, "prtimes")

		if !publishedTime.InRange(8) {
			break
		}

		href, _ := nodeDoc.Find("a").Attr("href")
		url := fmt.Sprintf("https://prtimes.jp%s", href)
		title := strings.TrimSpace(nodeDoc.Find("h3").Text())

		if strings.HasSuffix(title, "...") {
			articleDoc, err := getDoc(url)
			if err == nil {
				title = strings.Split(articleDoc.Find("title").Text(), " | ")[0]
			}
		}

		newsList = append(newsList, News{
			Title:         title,
			PublishedTime: publishedTime,
			Url:           url,
			Mention:       keyword,
		})
	}

	return newsList, nil
}

func getPanora(keyword string) ([]News, error) {
	var newsList []News

	var nodes []*html.Node

	doc, err := getDoc(fmt.Sprintf("https://panora.tokyo/?s=%s", url.QueryEscape(keyword)))
	if err != nil {
		return []News{}, err
	}
	nodes = append(nodes, doc.Find("article").Nodes[:3]...)

	doc, err = getDoc(fmt.Sprintf("https://panora.tokyo/archives/tag/%s", url.QueryEscape(keyword)))
	if err != nil {
		return []News{}, err
	}
	nodes = append(nodes, doc.Find("article").Nodes[:3]...)

	for _, node := range nodes {
		nodeDoc := goquery.NewDocumentFromNode(node)

		t, _ := nodeDoc.Find("time").Attr("datetime")
		publishedTime := stringToTime(t, "panora")

		if !publishedTime.InRange(8) {
			continue
		}

		url, _ := nodeDoc.Find("a").Attr("href")
		title := nodeDoc.Find("h2").Find("a").Text()

		news := News{
			Title:         title,
			PublishedTime: publishedTime,
			Url:           url,
			Mention:       keyword,
		}

		if isExist(newsList, news) {
			continue
		}

		newsList = append(newsList, news)
	}

	return newsList, nil
}

func getRealsound(keyword string) ([]News, error) {
	var newsList []News

	var nodes []*html.Node

	doc, err := getDoc(fmt.Sprintf("https://realsound.jp/?s=%s", url.QueryEscape(keyword)))
	if err != nil {
		return []News{}, err
	}
	nodes = append(nodes, doc.Find("article").Nodes[:3]...)

	if keyword == "湊あくあ" {
		doc, err = getDoc(fmt.Sprintf("https://realsound.jp/tag/%s", url.QueryEscape(keyword)))
		if err != nil {
			return []News{}, err
		}
		nodes = append(nodes, doc.Find("article").Nodes[:3]...)
	}

	for _, node := range nodes {
		nodeDoc := goquery.NewDocumentFromNode(node)

		t, _ := nodeDoc.Find("time").Attr("datetime")
		publishedTime := stringToTime(t, "realsound")

		if !publishedTime.InRange(8) {
			continue
		}

		url, _ := nodeDoc.Find("a").Attr("href")
		title := nodeDoc.Find("h3").Find("a").Text()

		news := News{
			Title:         title,
			PublishedTime: publishedTime,
			Url:           url,
			Mention:       keyword,
		}

		if isExist(newsList, news) {
			continue
		}

		newsList = append(newsList, news)
	}

	return newsList, nil
}

func getDoc(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &goquery.Document{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &goquery.Document{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return &goquery.Document{}, fmt.Errorf("HTTP request failed with status code: %d\n%s", resp.StatusCode, string(body))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return &goquery.Document{}, err
	}

	return doc, nil
}

func stringToTime(s, d string) Time {
	var t time.Time
	var layout string
	var err error

	if d == "realsound" {
		layout = "2006-01-02T15:04"
	} else if d == "prtimes" {
		layout = "2006-01-02T15:04:05-0700"
	}

	t, err = time.Parse(layout, s)
	if err != nil {
		return Time{}
	}

	return Time(t)
}

func isExist(list []News, target News) bool {
	for _, element := range list {
		if element == target {
			return true
		}
	}

	return false
}

func (news News) Map() map[string]any {
	newsMap := map[string]any{
		"Title":         news.Title,
		"PublishedTime": news.PublishedTime.UTC().String(),
		"Url":           news.Url,
		"Mention":       news.Mention,
	}

	return newsMap
}
