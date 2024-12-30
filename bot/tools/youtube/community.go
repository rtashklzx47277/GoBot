package youtube

import (
	"GoBot/tools"
	"fmt"
	"net/url"
	"strings"
)

func GetCommunity(channelId string) ([]Post, error) {
	url := fmt.Sprintf("https://www.youtube.com/channel/%s/community", channelId)
	reader, err := tools.Get(url).AddCookie("__Secure-3PSID", secure_3PSID).AddCookie("__Secure-3PSIDTS", secure_3PSIDTS).Do()
	if err != nil {
		return []Post{}, err
	}

	data, err := tools.ToString(reader)
	if err != nil {
		return []Post{}, err
	}

	match, ok := tools.Regexp(data, `ytInitialData = (.+?);\s*<\/script>`)
	if !ok {
		return []Post{}, fmt.Errorf("failed to get ytInitialData!\n%w", err)
	}

	jsonData, err := tools.StringToJson(match)
	if err != nil {
		return []Post{}, err
	}

	var posts []Post

	tab, ok := getTab(jsonData.Get("contents").Get("twoColumnBrowseResultsRenderer").Get("tabs"), "Community", "社群")
	if !ok {
		return []Post{}, nil
	}

	for _, item := range tab.Get("tabRenderer").Get("content").Get("sectionListRenderer").Get("contents").Index(0).Get("itemSectionRenderer").Get("contents").JsonArray() {
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

func getTab(item *tools.Json, target ...string) (*tools.Json, bool) {
	for _, tab := range item.JsonArray() {
		if tab.Get("tabRenderer").Get("selected").Bool() && tools.IsContain(target, tab.Get("tabRenderer").Get("title").String()) {
			return tab, true
		}
	}

	return &tools.Json{}, false
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
		renderer.Images = append(renderer.Images, getThumbnail(attachment.Get("backstageImageRenderer").Get("image")))
	} else if attachment.Exist("postMultiImageRenderer") {
		renderer.Type = "Image"

		for _, image := range attachment.Get("postMultiImageRenderer").Get("images").JsonArray() {
			renderer.Images = append(renderer.Images, getThumbnail(image.Get("backstageImageRenderer").Get("image")))
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
	if !item.Exist("thumbnails") {
		return ""
	}

	url := item.Get("thumbnails").Index(-1).Get("url").String()

	if strings.Contains(url, "=s") {
		return strings.Split(url, "=s")[0] + "=s0"
	}

	return strings.Split(url, "=w")[0] + "=w0"
}
