package youtube

import (
	"GoBot/tools"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func GetCommunity(channelId string) ([]Post, error) {
	url := fmt.Sprintf("https://www.youtube.com/channel/%s/community", channelId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []Post{}, err
	}

	// __Secure-3PSID
	// __Secure-3PSIDTS

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []Post{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []Post{}, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []Post{}, err
	}

	data := tools.Regexp(string(body), `ytInitialData = (.+);\s*<\/script>`, 1)[0][1]

	var jsonData tools.Json

	err = json.Unmarshal([]byte(data), &jsonData.Data)
	if err != nil {
		return []Post{}, err
	}

	var posts []Post

	for _, item := range getTab(jsonData.Get("contents").Get("twoColumnBrowseResultsRenderer").Get("tabs")).Get("tabRenderer").Get("content").Get("sectionListRenderer").Get("contents").Index(0).Get("itemSectionRenderer").Get("contents").JsonArray() {
		if !item.Exist("backstagePostThreadRenderer") {
			break
		}

		item = item.Get("backstagePostThreadRenderer").Get("post").Get("backstagePostRenderer")

		posts = append(posts, Post{
			Id:       item.Get("postId").String(),
			Url:      fmt.Sprintf("https://www.youtube.com/post/%s", item.Get("postId").String()),
			Text:     getContentText(item.Get("contentText")),
			Member:   item.Exist("sponsorsOnlyBadge"),
			Renderer: getRenderer(item),
		})
	}

	return posts, nil
}

func getTab(js *tools.Json) *tools.Json {
	for _, tab := range js.JsonArray() {
		if tab.Get("tabRenderer").Get("selected").Bool() {
			return tab
		}
	}

	return &tools.Json{}
}

func getContentText(js *tools.Json) string {
	var text string

	for _, run := range js.Get("runs").JsonArray() {
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

func getRenderer(js *tools.Json) Renderer {
	if !js.Exist("backstageAttachment") {
		return Renderer{Type: "None"}
	}

	attachment := js.Get("backstageAttachment")

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
		renderer.Video.Url = fmt.Sprintf("https://www.youtube.com/watch?v=%s", renderer.Video.Id)
	} else if attachment.Exist("playlistRenderer") {
		renderer.Type = "Playlist"
		renderer.Playlist.Id = attachment.Get("playlistRenderer").Get("playlistId").String()
		renderer.Playlist.Url = fmt.Sprintf("https://www.youtube.com/playlist?list=%s", renderer.Playlist.Id)
	} else if attachment.Exist("pollRenderer") {
		renderer.Type = "Poll"
		renderer.Choices = getPoll(js)
	} else if attachment.Exist("quizRenderer") {
		renderer.Type = "Quiz"
		renderer.Choices = getQuiz(js)
	}

	return renderer
}

func getPoll(js *tools.Json) []Choice {
	var choices []Choice

	if !js.Get("backstageAttachment").Exist("pollRenderer") {
		return choices
	}

	for _, choice := range js.Get("backstageAttachment").Get("pollRenderer").Get("choices").JsonArray() {
		choices = append(choices, Choice{
			Text:  choice.Get("text").Get("runs").Index(0).Get("text").String(),
			Image: getThumbnail(choice),
		})
	}

	return choices
}

func getQuiz(js *tools.Json) []Choice {
	var choices []Choice

	if !js.Get("backstageAttachment").Exist("quizRenderer") {
		return choices
	}

	for _, choice := range js.Get("backstageAttachment").Get("quizRenderer").Get("choices").JsonArray() {
		choices = append(choices, Choice{
			Text:      choice.Get("text").Get("runs").Index(0).Get("text").String(),
			Image:     choice.Get("explanation").Get("runs").Index(0).Get("text").String(),
			isCorrect: choice.Get("isCorrect").Bool(),
		})
	}

	return choices
}

func getThumbnail(js *tools.Json) string {
	if !js.Exist("image") {
		return ""
	}

	return js.Get("image").Get("thumbnails").Index(-1).Get("url").Split("=s")[0]
}
