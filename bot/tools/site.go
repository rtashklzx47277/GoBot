package tools

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Article struct {
	Id        string
	Title     string
	Date      Time
	Content   string
	Thumbnail string
	Category  string
	Type      string
	Url       string
	From      string
}

func GetHP(lastestArticleId string) ([]Article, error) {
	reader, err := Get("https://yuuki-sakuna.com/news-list").Do()
	if err != nil {
		return []Article{}, err
	}

	doc, err := ToDocument(reader)
	if err != nil {
		return []Article{}, err
	}

	nodes := doc.Find(".pc_only li.p-postList__item").Nodes
	for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
		nodes[i], nodes[j] = nodes[j], nodes[i]
	}

	var articles []Article

	for _, node := range nodes[max(0, len(nodes)-5):] {
		nodeDoc := goquery.NewDocumentFromNode(node)

		url, _ := nodeDoc.Find(".p-postList__link").Attr("href")
		urlSet := strings.Split(url, "/")
		articleId := urlSet[len(urlSet)-2]

		if articleId == lastestArticleId {
			break
		}

		date, _ := nodeDoc.Find(".c-postTimes__posted").Attr("datetime")
		parsedDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			return articles, err
		}

		thumbnails, _ := nodeDoc.Find(".c-postThumb__figure").Find("img").Attr("data-srcset")
		thumbnailSet := strings.Split(thumbnails, ",")
		thumbnail := strings.Split(thumbnailSet[len(thumbnailSet)-1], " ")[1]

		reader, err := Get(url).Do()
		if err != nil {
			return articles, err
		}

		doc, err := ToDocument(reader)
		if err != nil {
			return articles, err
		}

		content := strings.TrimSpace(doc.Find("#main_content .post_content").Text())

		article := Article{
			Id:        articleId,
			Title:     nodeDoc.Find(".p-postList__title").Text(),
			Date:      Time(parsedDate),
			Content:   content,
			Thumbnail: thumbnail,
			Category:  nodeDoc.Find(".icon-folder").Text(),
			Type:      "",
			Url:       url,
			From:      "HP",
		}

		articles = append(articles, article)
	}

	return reverse(articles), nil
}

func GetRadio(lastestArticleId string) ([]Article, error) {
	url := "https://api.qlover.jp/fc/fanclub_sites/690/article_themes/news/articles?per_page=5&sort=published_at_desc"
	reader, err := Get(url).AddHeader("fc_use_device", "null").Do()
	if err != nil {
		return []Article{}, err
	}

	data, err := ToJson(reader)
	if err != nil {
		return []Article{}, err
	}

	var articles []Article

	for _, article := range data.Get("data").Get("article_theme").Get("articles").Get("list").JsonArray() {
		articleId := article.Get("article_code").String()
		if articleId == lastestArticleId {
			break
		}

		url := fmt.Sprintf("https://api.qlover.jp/fc/fanclub_sites/690/article_themes/news/articles/%s", articleId)
		reader, err := Get(url).AddHeader("fc_use_device", "null").Do()
		if err != nil {
			return []Article{}, err
		}

		data, err := ToJson(reader)
		if err != nil {
			return []Article{}, err
		}

		content := data.Get("data").Get("article").Get("article").Get("contents").String()

		var categories []string
		for _, category := range article.Get("article_article_categories").JsonArray() {
			categories = append(categories, category.Get("article_category").Get("category_name").String())
		}

		thumbnail := article.Get("thumbnail_url").String()
		if thumbnail == "null" {
			thumbnail = ""
		}

		article := Article{
			Id:        articleId,
			Title:     article.Get("article_title").String(),
			Date:      article.Get("publish_at").Time(),
			Content:   content,
			Thumbnail: thumbnail,
			Category:  strings.Join(categories, " "),
			Type:      article.Get("article_authorization_type").Get("authorization_name").String(),
			Url:       fmt.Sprintf("https://qlover.jp/sakuna/articles/news/%s", articleId),
			From:      "Radio",
		}

		articles = append(articles, article)
	}

	return reverse(articles), nil
}

func (article Article) Map() map[string]any {
	articleMap := map[string]any{
		"Id":       article.Id,
		"Title":    article.Title,
		"Date":     article.Date.String(),
		"Content":  article.Content,
		"Category": article.Category,
		"Type":     article.Type,
		"`From`":   article.From,
	}

	return articleMap
}

func reverse(articles []Article) []Article {
	for i, j := 0, len(articles)-1; i < j; i, j = i+1, j-1 {
		articles[i], articles[j] = articles[j], articles[i]
	}

	return articles
}
