package youtube

import (
	"GoBot/tools"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	secure_3PSID   = os.Getenv("SECURE_3PSID")
	secure_3PSIDTS = os.Getenv("SECURE_3PSIDTS")
)

func GetCommunity(channelId string) ([]Post, error) {
	url := fmt.Sprintf("https://www.youtube.com/channel/%s/community", channelId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []Post{}, fmt.Errorf("error occurred at line 21: %v", err)
	}

	req.AddCookie(&http.Cookie{Name: "__Secure-3PSID", Value: secure_3PSID})
	req.AddCookie(&http.Cookie{Name: "__Secure-3PSIDTS", Value: secure_3PSIDTS})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []Post{}, fmt.Errorf("error occurred at line 29: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []Post{}, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []Post{}, fmt.Errorf("error occurred at line 39: %v", err)
	}

	data := tools.Regexp(string(body), `ytInitialData = (.+?);\s*<\/script>`, 1)[0][1]

	var jsonData tools.Json

	err = json.Unmarshal([]byte(data), &jsonData.Data)
	if err != nil {
		return []Post{}, fmt.Errorf("error occurred at line 48: %v", err)
	}

	var posts []Post

	for _, item := range getTab(jsonData.Get("contents").Get("twoColumnBrowseResultsRenderer").Get("tabs")).Get("tabRenderer").Get("content").Get("sectionListRenderer").Get("contents").Index(0).Get("itemSectionRenderer").Get("contents").JsonArray() {
		if !item.Exist("backstagePostThreadRenderer") {
			break
		}

		item = item.Get("backstagePostThreadRenderer").Get("post").Get("backstagePostRenderer")

		postId := item.Get("postId").String()

		post := Post{
			Id:       postId,
			Url:      fmt.Sprintf("https://www.youtube.com/post/%s", postId),
			Text:     getContentText(item),
			Member:   item.Exist("sponsorsOnlyBadge"),
			Renderer: getRenderer(item),
		}

		post.Author.Id = channelId

		posts = append(posts, post)
	}

	return posts, nil
}

func getTab(item *tools.Json) *tools.Json {
	for _, tab := range item.JsonArray() {
		if tab.Get("tabRenderer").Get("selected").Bool() {
			return tab
		}
	}

	return &tools.Json{}
}

func getContentText(item *tools.Json) string {
	var text string

	for _, run := range item.Get("contentText").Get("runs").JsonArray() {
		if run.Exist("navigationEndpoint") {
			decodedUrl, err := url.QueryUnescape(run.Get("navigationEndpoint").Get("commandMetadata").Get("webCommandMetadata").Get("url").String())
			if err != nil {
				text += run.Get("text").String()
			}

			if strings.Contains(decodedUrl, "youtube.com/redirect") {
				text += strings.Split(decodedUrl, "q=")[1]
			} else {
				text += fmt.Sprintf("https://www.youtube.com%s", decodedUrl)
			}
		} else {
			text += run.Get("text").String()
		}
	}

	return text
}

func getRenderer(item *tools.Json) Renderer {
	if !item.Exist("backstageAttachment") {
		return Renderer{Type: "None"}
	}

	attachment := item.Get("backstageAttachment")

	var renderer Renderer

	if attachment.Exist("backstageImageRenderer") {
		renderer.Type = "Image"
		renderer.Images = append(renderer.Images, getThumbnail(attachment.Get("backstageImageRenderer")))
	} else if attachment.Exist("postMultiImageRenderer") {
		renderer.Type = "Image"

		for _, image := range attachment.Get("postMultiImageRenderer").Get("images").JsonArray() {
			renderer.Images = append(renderer.Images, getThumbnail(image.Get("backstageImageRenderer")))
		}
	} else if attachment.Exist("videoRenderer") {
		renderer.Type = "Video"
		renderer.Video.Id = attachment.Get("videoRenderer").Get("videoId").String()
	} else if attachment.Exist("playlistRenderer") {
		renderer.Type = "Playlist"
		renderer.Playlist.Id = attachment.Get("playlistRenderer").Get("playlistId").String()
	} else if attachment.Exist("pollRenderer") {
		renderer.Type = "Poll"
		renderer.Choices = getPoll(item)
	} else if attachment.Exist("quizRenderer") {
		renderer.Type = "Quiz"
		renderer.Choices = getQuiz(item)
	}

	return renderer
}

func getPoll(item *tools.Json) []Choice {
	var choices []Choice

	if !item.Get("backstageAttachment").Exist("pollRenderer") {
		return choices
	}

	for _, choice := range item.Get("backstageAttachment").Get("pollRenderer").Get("choices").JsonArray() {
		choices = append(choices, Choice{
			Type: "Poll",
			Text: choice.Get("text").Get("runs").Index(0).Get("text").String(),
		})
	}

	return choices
}

func getQuiz(item *tools.Json) []Choice {
	var choices []Choice

	if !item.Get("backstageAttachment").Exist("quizRenderer") {
		return choices
	}

	for _, choice := range item.Get("backstageAttachment").Get("quizRenderer").Get("choices").JsonArray() {
		choices = append(choices, Choice{
			Type:    "Quiz",
			Text:    choice.Get("text").Get("runs").Index(0).Get("text").String(),
			Correct: choice.Get("isCorrect").Bool(),
		})
	}

	return choices
}

func getThumbnail(item *tools.Json) string {
	if !item.Exist("image") {
		return ""
	}

	return item.Get("image").Get("thumbnails").Index(-1).Get("url").Split("=s")[0]
}
