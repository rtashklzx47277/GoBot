package youtube

import (
	"GoBot/tools"
	"fmt"
	"io"
	"net/http"
)

func GetCollab(channelId string) ([]string, error) {
	var videoIdList []string

	url := fmt.Sprintf("https://holodex.net/api/v2/channels/%s/collabs?type=stream,placeholder&include=clips,live_info&limit=3", channelId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []string{}, err
	}

	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Encoding", "deflate")
	req.Header.Set("Referer", fmt.Sprintf("https://holodex.net/channel/%s/collabs", channelId))
	req.Header.Set("User-Agent", tools.UserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return []string{}, fmt.Errorf("HTTP request failed with status code: %d\n%s", resp.StatusCode, string(body))
	}

	data, err := tools.ToJson(resp.Body)
	if err != nil {
		return []string{}, err
	}

	for _, video := range data.JsonArray() {
		videoIdList = append(videoIdList, video.Get("id").String())
	}

	return videoIdList, nil
}
