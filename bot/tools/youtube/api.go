package youtube

import (
	"GoBot/tools"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	apiKeyList = map[int]string{
		1: os.Getenv("YOUTUBE_API_KEY_1"),
		2: os.Getenv("YOUTUBE_API_KEY_2"),
		3: os.Getenv("YOUTUBE_API_KEY_3"),
		4: os.Getenv("YOUTUBE_API_KEY_4"),
		5: os.Getenv("YOUTUBE_API_KEY_5"),
		6: os.Getenv("YOUTUBE_API_KEY_6"),
		7: os.Getenv("YOUTUBE_API_KEY_7"),
		8: os.Getenv("YOUTUBE_API_KEY_8"),
	}
	checkList = map[string]string{
		"UCJFZiqLMntJufDCHc6bQixg": "ホロライブ",
		"UCp6993wxpyDPHUpavwDFqgg": "ときのそら",
		"UCDqI2jOz0weumE8s7paEk6g": "ロボ子さん",
		"UC5CwaMl1eIgY8h02uZw7u8A": "星街すいせい",
		"UC-hM6YJuNYVAmUWxeIr9FeA": "さくらみこ",
		"UC0TXe_LYZ4scaW2XMyi5_kw": "AZKi",
		"UCD8HOxPs4Xvsm8H0ZxXGiBw": "夜空メル",
		"UCdn5BQ06XqgXoAxIhbqw5Rg": "白上フブキ",
		"UCQ0UDLQCjY0rmuxCDE38FGg": "夏色まつり",
		"UCFTLzh12_nrtzqBPsTCqenA": "アキ・ローゼンタール",
		"UC1CfXB_kRs3C-zaeTG3oGyg": "赤井はあと",
		"UC1opHUrw8rvnsadT-iGp7Cg": "湊あくあ",
		"UCXTpFs_3PqI41qX2d9tL2Rw": "紫咲シオン",
		"UC7fk0CB07ly8oSl0aqKkqFg": "百鬼あやめ",
		"UC1suqwovbL1kzsoaZgFZLKg": "癒月ちょこ",
		"UCvzGlP9oQwU--Y0r9id_jnA": "大空スバル",
		"UCp-5t9SrOQwXMU7iIjQfARg": "大神ミオ",
		"UCvaTdHTWBGv3MKj3KVqJVCw": "猫又おかゆ",
		"UChAnqc_AY5_I3Px5dig3X1Q": "戌神ころね",
		"UC1DCedRgGHBdm81E1llLhOQ": "兎田ぺこら",
		"UCl_gCybOJRIgOXw6Qb4qJzQ": "潤羽るしあ",
		"UCvInZx9h3jC2JzsIzoOebWg": "不知火フレア",
		"UCdyqAaZDKHXg4Ahi7VENThQ": "白銀ノエル",
		"UCCzUftO8KOVkV4wQG1vkUvg": "宝鐘マリン",
		"UCZlDXzGoo7d44bwdNObFacg": "天音かなた",
		"UCS9uQI-jC3DE0L4IpXyvr6w": "桐生ココ",
		"UCqm3BQLlJfvkTsX_hvm0UmA": "角巻わため",
		"UC1uv2Oq6kNxgATlCiez59hw": "常闇トワ",
		"UCa9Y57gfeY0Zro_noHRVrnw": "姫森ルーナ",
		"UCFKOVgVbGmX65RxO3EtH3iw": "雪花ラミィ",
		"UCAWSyEs_Io8MtpY3m-zqILA": "桃鈴ねね",
		"UCUKD-uaobj9jiqB-VXt71mA": "獅白ぼたん",
		"UCgZuwn-O7Szh9cAgHqJ6vjw": "魔乃アロエ",
		"UCK9V2B22uJYu3N7eR_BT9QA": "尾丸ポルカ",
		"UCENwRMx5Yh42zWpzURebzTw": "ラプラス・ダークネス",
		"UCs9_O1tRPMQTHQ-N_L6FU2g": "鷹嶺ルイ",
		"UC6eWCld0KwmyHFbAqK3V-Rw": "博衣こより",
		"UCIBY1ollUsauvVi4hW4cumw": "沙花叉クロヱ",
		"UC_vMYWcDjmfdpH6r4TTn1MQ": "風真いろは",
		"UCMGfV7TVTmHhEErVJg1oHBQ": "火威青",
		"UCWQtYtq9EOB4-I5P-3fh8lA": "音乃瀬奏",
		"UCtyWhCj3AqKh2dXctLkDtng": "一条莉々華",
		"UCdXAk5MpyLD8594lm_OvtGQ": "儒烏風亭らでん",
		"UC1iA6_NT4mtAcIII6ygrvCw": "轟はじめ",
		"UCOyYb1c43VlX9rc_lT6NKQw": "Ayunda Risu",
		"UCP0BspO_AMEe3aQqqpo89Dg": "Moona Hoshinova",
		"UCAoy6rzhSf4ydcYjJw3WoVg": "Airani Iofifteen",
		"UCYz_5n-uDuChHtLo7My1HnQ": "Kureiji Ollie",
		"UC727SQYUvx5pDDGQpTICNWg": "Anya Melfissa",
		"UChgTyjG-pdNvxxhdsXfHQ5Q": "Pavolia Reine",
		"UCTvHWSfBZgtxE4sILOaurIQ": "Vestia Zeta",
		"UCZLZ8Jjx_RN2CXloOmgTHVg": "Kaela Kovalskia",
		"UCjLEmnpCNeisMxy134KPwWw": "Kobo Kanaeru",
		"UCL_qhgtOy0dy1Agp8vkySQg": "Mori Calliope",
		"UCHsx4Hqa-1ORjQTh9TYDhww": "Takanashi Kiara",
		"UCMwGHR0BTZuLsmjY_NT5Pwg": "Ninomae Ina'nis",
		"UCoSrY_IQQVpmIRZ9Xf-y93g": "Gawr Gura",
		"UCyl1z3jo3XHR1riLFKG5UAg": "Watson Amelia",
		"UCsUj0dszADCGbF3gNrQEuSQ": "Tsukumo Sana",
		"UCO_aKKYxn4tvrqPjcTzZ6EQ": "Ceres Fauna",
		"UCmbs8T6MWqUHP1tIQvSgKrg": "Ouro Kronii",
		"UC3n5uGu18FoCy23ggWWp8tA": "Nanashi Mumei",
		"UCgmPnx-EEeOrZSg5Tiw7ZRQ": "Hakos Baelz",
		"UC8rcEBzJSleTkf_-agPM20g": "IRyS",
		"UC9p_lqQ0FEDz327Vgf5JwqA": "Koseki Bijou",
		"UCgnfPPb9JI3e9A4cXHnWbyg": "Shiori Novella",
		"UC_sFNM0z0MWm9A6WlKPuMMg": "Nerissa Ravencroft",
		"UCt9H_RpQzhxzlyBxFqrdHqA": "FUWAMOCO",
		"UCWCc8tO-uUl_7SJXIKJACMw": "神楽めあ",
		"UC8NZiqKx6fsDT3AVcMiVFyA": "犬山たまき",
		"UC_4tXjqecqox5Uc05ncxpxg": "椎名唯華",
		"UCoztvTULBYd3WmStqYeoHcA": "笹木咲",
		"UC9V3Y3_uzU5e-usObb6IE1w": "星川サラ",
		"UC9EjSJ8pvxtvPdxLOElv73w": "魔界ノりりむ",
	}
)

func getData(path string) (*tools.Json, error) {
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/%s&key=%s", path, apiKeyList[time.Now().Hour()/3+1])

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &tools.Json{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &tools.Json{}, err
	}
	defer resp.Body.Close()

	data, err := tools.ToJson(resp.Body)
	if err != nil {
		return &tools.Json{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return &tools.Json{}, fmt.Errorf("HTTP request failed with status code: %d\n%s", resp.StatusCode, string(body))
	}

	return data, nil
}

func GetChannel(channelId string) (Channel, error) {
	path := fmt.Sprintf("channels?part=brandingSettings,snippet,statistics&id=%s", channelId)
	data, err := getData(path)
	if err != nil {
		return Channel{}, err
	}

	if !data.Exist("items") {
		return Channel{}, errors.New("fail to get data")
	}

	item := data.Get("items").Index(0)
	channel := Channel{
		Id:              item.Get("id").String(),
		CustomId:        item.Get("snippet").Get("customUrl").Slice(1, -1),
		Url:             fmt.Sprintf("https://www.youtube.com/channel/%s", item.Get("id").String()),
		Title:           item.Get("snippet").Get("title").String(),
		Description:     item.Get("snippet").Get("description").String(),
		Icon:            item.Get("snippet").Get("thumbnails").Image(),
		Banner:          item.Get("brandingSettings").Get("image").Get("bannerExternalUrl").String(),
		SubscriberCount: item.Get("statistics").Get("subscriberCount").Int(),
		ViewCount:       item.Get("statistics").Get("viewCount").Int(),
	}

	return channel, nil
}

func GetChannelSections(channelId string) ([]map[string]any, error) {
	path := fmt.Sprintf("channelSections?part=contentDetails,snippet&channelId=%s", channelId)
	data, err := getData(path)
	if err != nil {
		return []map[string]any{}, err
	}

	var sections []map[string]any

	for _, item := range data.Get("items").JsonArray() {
		var section map[string]any
		sectionType := item.Get("snippet").Get("type").String()

		if sectionType == "channelsectiontypeundefined" {
			continue
		} else if sectionType == "multiplechannels" || sectionType == "singleplaylist" || sectionType == "multipleplaylists" {
			var sectionContent []any

			if sectionType == "multiplechannels" {
				sectionContent = item.Get("contentDetails").Get("channels").Array()
			} else {
				sectionContent = item.Get("contentDetails").Get("playlists").Array()
			}

			section = map[string]any{
				"Type":    sectionType,
				"Content": sectionContent,
			}
		} else {
			section = map[string]any{
				"Type": sectionType,
			}
		}

		sections = append(sections, section)
	}

	return sections, nil
}

func GetPlaylists(channelId string) ([]Playlist, error) {
	var playlists []Playlist
	var pageToken string

	for {
		path := fmt.Sprintf("playlists?part=snippet,status&channelId=%s&maxResults=50&pageToken=%s", channelId, pageToken)
		data, err := getData(path)
		if err != nil {
			return []Playlist{}, err
		}

		if !data.Exist("items") {
			return []Playlist{}, errors.New("fail to get data")
		}

		pageToken = data.Get("nextPageToken").String()

		for _, item := range data.Get("items").JsonArray() {
			playlist := Playlist{
				Id:          item.Get("id").String(),
				Url:         fmt.Sprintf("https://www.youtube.com/playlist?list=%s", item.Get("id").String()),
				Title:       item.Get("snippet").Get("title").String(),
				Description: item.Get("snippet").Get("description").String(),
				Thumbnail:   item.Get("snippet").Get("thumbnails").Image(),
			}
			playlist.Author.Id = channelId
			playlists = append(playlists, playlist)
		}

		if pageToken == "" {
			break
		}
	}

	return playlists, nil
}

func GetPlaylistItems(playlistId string, num int) ([]Video, error) {
	var videos []Video
	var pageToken string

	for {
		path := fmt.Sprintf("playlistItems?part=snippet&playlistId=%s&maxResults=%d&pageToken=%s", playlistId, num, pageToken)
		data, err := getData(path)
		if err != nil {
			return []Video{}, err
		}

		if !data.Exist("items") {
			return []Video{}, errors.New("fail to get data")
		}

		pageToken = data.Get("nextPageToken").String()

		for _, item := range data.Get("items").JsonArray() {
			video := Video{
				Id:        item.Get("snippet").Get("resourceId").Get("videoId").String(),
				Url:       fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.Get("snippet").Get("resourceId").Get("videoId").String()),
				Title:     item.Get("snippet").Get("title").String(),
				Thumbnail: item.Get("snippet").Get("thumbnails").Image(),
			}

			videos = append(videos, video)
		}

		if pageToken == "" || num != 50 {
			break
		}
	}

	return videos, nil
}

func GetVideo(videoId string) (Video, error) {
	path := fmt.Sprintf("videos?part=contentDetails,liveStreamingDetails,snippet,statistics&id=%s", videoId)
	data, err := getData(path)
	if err != nil {
		return Video{}, err
	}

	if !data.Exist("items") {
		return Video{}, errors.New("fail to get data")
	}

	var item *tools.Json

	if len(data.Get("items").JsonArray()) != 0 {
		item = data.Get("items").Index(0)
	} else {
		item = nil
	}

	return getVideoStruct(item, videoId), nil
}

func GetVideos(videoIds []string) ([]Video, error) {
	var videos []Video
	length := len(videoIds)

	for n := 0; n < length; n += 50 {
		end := n + 50

		if end > length {
			end = length
		}

		path := fmt.Sprintf("videos?part=contentDetails,liveStreamingDetails,snippet,statistics&id=%s", strings.Join(videoIds[n:end], ","))
		data, err := getData(path)
		if err != nil {
			return []Video{}, err
		}

		if !data.Exist("items") {
			return []Video{}, errors.New("fail to get data")
		}

		dataList := data.Get("items").JsonArray()

		for _, videoId := range videoIds[n:end] {
			var item *tools.Json

			if len(dataList) != 0 && videoId == dataList[0].Get("id").String() {
				item = dataList[0]
				dataList = dataList[1:]
			} else {
				item = nil
			}

			videos = append(videos, getVideoStruct(item, videoId))
		}
	}

	return videos, nil
}

func GetComments(target, Id string) ([]Comment, error) {
	var comments []Comment
	var request, filter, pageToken string

	if target == "channel" {
		request, filter = "commentThreads", "allThreadsRelatedToChannelId"
	} else if target == "video" {
		request, filter = "commentThreads", "videoId"
	} else if target == "reply" {
		request, filter = "comments", "parentId"
	}

	for {
		path := fmt.Sprintf("%s?part=snippet&%s=%s&maxResults=100&textFormat=plainText&pageToken=%s", request, filter, Id, pageToken)
		data, err := getData(path)
		if err != nil {
			return []Comment{}, err
		}

		if !data.Exist("items") {
			return []Comment{}, errors.New("fail to get data")
		}

		pageToken = data.Get("nextPageToken").String()
		dataList := data.Get("items").JsonArray()

		if len(dataList) == 0 {
			pageToken = ""
		}

		for _, item := range dataList {
			var snippet *tools.Json

			if target == "channel" || target == "video" {
				snippet = item.Get("snippet").Get("topLevelComment").Get("snippet")
			} else if target == "reply" {
				snippet = item.Get("snippet")
			}

			commentTime := snippet.Get("publishedAt").Time()
			authorId := snippet.Get("authorChannelId").Get("value").String()

			if !commentTime.InRange(1) {
				pageToken = ""
				break
			} else if _, ok := checkList[authorId]; ok {
				var id string

				if target == "channel" || target == "video" {
					id = item.Get("id").String()
				} else if target == "reply" {
					id = item.Get("id").Split(".")[1]
				}

				comment := Comment{
					Id:            id,
					Text:          snippet.Get("textDisplay").String(),
					PublishedTime: commentTime,
					UpdatedTime:   snippet.Get("updatedAt").Time(),
				}

				comment.Author.Id = authorId
				comment.Author.Url = fmt.Sprintf("https://www.youtube.com/channel/%s", authorId)

				if target == "channel" || target == "video" {
					comment.Video.Id = snippet.Get("videoId").String()
					comment.Video.Url = fmt.Sprintf("https://www.youtube.com/watch?v=%s", comment.Video.Id)
				} else if target == "reply" {
					comment.ParentId = Id
				}

				comments = append(comments, comment)
			}
		}

		if pageToken == "" {
			break
		}
	}

	return comments, nil
}
