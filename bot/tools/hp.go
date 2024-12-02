package tools

import (
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Article struct {
	Title     string
	Date      Time
	Content   string
	Thumbnail string
	Category  string
	Url       string
}

func GetHP() ([]Article, error) {
	reader, err := Get("https://yuuki-sakuna.com/news-list/").Do()
	if err != nil {
		return []Article{}, err
	}

	doc, err := ToDocument(reader)
	if err != nil {
		return []Article{}, err
	}

	var articles []Article

	for _, node := range doc.Find(".pc_only li.p-postList__item").Nodes {
		nodeDoc := goquery.NewDocumentFromNode(node)

		date, _ := nodeDoc.Find(".c-postTimes__posted").Attr("datetime")
		parsedDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			return articles, err
		}

		thumbnails, _ := nodeDoc.Find(".c-postThumb__figure").Find("img").Attr("data-srcset")
		thumbnailSet := strings.Split(thumbnails, ",")
		thumbnail := strings.Split(thumbnailSet[len(thumbnailSet)-1], " ")[1]

		url, _ := nodeDoc.Find(".p-postList__link").Attr("href")

		reader, err := Get(url).Do()
		if err != nil {
			return articles, err
		}

		articleDoc, err := ToDocument(reader)
		if err != nil {
			return articles, err
		}

		article := Article{
			Title:     nodeDoc.Find(".p-postList__title").Text(),
			Date:      Time(parsedDate),
			Content:   strings.TrimSpace(articleDoc.Find("#main_content .post_content").Text()),
			Thumbnail: thumbnail,
			Category:  nodeDoc.Find(".icon-folder").Text(),
			Url:       url,
		}

		articles = append(articles, article)
	}

	return articles, nil
}
