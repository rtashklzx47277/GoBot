package youtube

import (
	"GoBot/tools"
	"fmt"
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
