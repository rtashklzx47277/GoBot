package main

import (
	"GoBot/tools"
	"GoBot/tools/youtube"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var clientVersion = "2.20240620.05.00"

type Message struct {
	Id       string
	VideoId  string
	AuthorId string
	Time     tools.Time
	Badge    string
	Amount   string
	Text     string
}

func LiveChat(video youtube.Video, channel youtube.Channel) {
	apiKey, continuation, err := getParameters(video.Id)
	if err != nil {
		return
	}

	messageIdList = append(messageIdList, db.Distinct("Message", video.Id)...)

	count := 0

	for count < 5 {
		data, err := getChatData(apiKey, continuation)
		if err != nil {
			continue
		}

		if !data.Exist("continuationContents") {
			count++
			fmt.Println("Can't find continuationContents!")
			continue
		}

		continuations := data.Get("continuationContents").Get("liveChatContinuation").Get("continuations").Index(0)

		if continuations.Exist("timedContinuationData") {
			continuation = continuations.Get("timedContinuationData").Get("continuation").String()
		} else if continuations.Exist("invalidationContinuationData") {
			continuation = continuations.Get("invalidationContinuationData").Get("continuation").String()
		}

		for _, action := range data.Get("continuationContents").Get("liveChatContinuation").Get("actions").JsonArray() {
			getMessageData(action, video, channel)
		}
	}
}

func getParameters(videoId string) (string, string, error) {
	url := fmt.Sprintf("https://www.youtube.com/live_chat?v=%s", videoId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Set("User-Agent", youtube.UserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return "", "", fmt.Errorf("HTTP request failed with status code: %d\n%s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	apiKey, err := getApiKey(body)
	if err != nil {
		return "", "", err
	}

	continuation, err := getContinuation(body)
	if err != nil {
		return "", "", err
	}

	return apiKey, continuation, nil
}

func getApiKey(body []byte) (string, error) {
	match := tools.Regexp(string(body), `"INNERTUBE_API_KEY":"([A-z0-9-]*)`, 1)

	if len(match) == 0 {
		return "", errors.New("fail to get apiKey")
	}

	return match[0][1], nil
}

func getContinuation(body []byte) (string, error) {
	match := tools.Regexp(string(body), `"continuation":"([A-z0-9-%]*)`, -1)

	if len(match) == 0 {
		return "", errors.New("fail to get continuation")
	}

	return match[len(match)-1][1], nil
}

func getChatData(apiKey, continuation string) (*tools.Json, error) {
	url := fmt.Sprintf("https://www.youtube.com/youtubei/v1/live_chat/get_live_chat?key=%s", apiKey)

	payload, err := getPayload(continuation)
	if err != nil {
		return &tools.Json{}, err
	}

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return &tools.Json{}, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &tools.Json{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return &tools.Json{}, fmt.Errorf("HTTP request failed with status code: %d\n%s", resp.StatusCode, string(body))
	}

	data, err := tools.ToJson(resp.Body)
	if err != nil {
		return &tools.Json{}, err
	}

	return data, nil
}

func getPayload(continuation string) (io.Reader, error) {
	payload := map[string]any{
		"context": map[string]any{
			"client": map[string]string{
				"clientName":    "WEB",
				"clientVersion": clientVersion,
			},
		},
		"continuation": continuation,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(payloadBytes)), nil
}

func getMessageData(action *tools.Json, video youtube.Video, channel youtube.Channel) {
	if action.Exist("addChatItemAction") {
		item := action.Get("addChatItemAction").Get("item")

		if item.Exist("liveChatTextMessageRenderer") {
			rendererProcessor(item.Get("liveChatTextMessageRenderer"), "TextMessage", video, channel)
		} else if item.Exist("liveChatPaidMessageRenderer") {
			rendererProcessor(item.Get("liveChatPaidMessageRenderer"), "PaidMessage", video, channel)
		} else if item.Exist("liveChatPaidStickerRenderer") {
			rendererProcessor(item.Get("liveChatPaidStickerRenderer"), "PaidSticker", video, channel)
		} else if item.Exist("liveChatMembershipItemRenderer") {
			rendererProcessor(item.Get("liveChatMembershipItemRenderer"), "Membership", video, channel)
		} else if item.Exist("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer") {
			rendererProcessor(item.Get("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer"), "GiftSend", video, channel)
		} else if item.Exist("liveChatSponsorshipsGiftRedemptionAnnouncementRenderer") {
			rendererProcessor(item.Get("liveChatSponsorshipsGiftRedemptionAnnouncementRenderer"), "GiftReceive", video, channel)
		} else if item.Exist("liveChatModeChangeMessageRenderer") {
			liveChatSetting(item.Get("liveChatModeChangeMessageRenderer"))
		} else if item.Exist("liveChatViewerEngagementMessageRenderer") {
		} else if item.Exist("liveChatPlaceholderItemRenderer") {
		} else {
			fmt.Println("Error getting renderer from addChatItemAction!")
			fmt.Println(toJSON(item))
		}
	} else if action.Exist("addLiveChatTickerItemAction") {
		item := action.Get("addLiveChatTickerItemAction").Get("item")

		if item.Exist("liveChatTickerPaidMessageItemRenderer") {
			rendererProcessor(item.Get("liveChatTickerPaidMessageItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatPaidMessageRenderer"), "PaidMessage", video, channel)
		} else if item.Exist("liveChatTickerPaidStickerItemRenderer") {
			rendererProcessor(item.Get("liveChatTickerPaidStickerItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatPaidStickerRenderer"), "PaidSticker", video, channel)
		} else if item.Exist("liveChatTickerSponsorItemRenderer") {
			if item.Get("liveChatTickerSponsorItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Exist("liveChatMembershipItemRenderer") {
				rendererProcessor(item.Get("liveChatTickerSponsorItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatMembershipItemRenderer"), "Membership", video, channel)
			} else if item.Get("liveChatTickerSponsorItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Exist("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer") {
				rendererProcessor(item.Get("liveChatTickerSponsorItemRenderer"), "GiftSend", video, channel)
			} else {
				fmt.Println("Error getting renderer from liveChatTickerSponsorItemRenderer!")
				fmt.Println(toJSON(item))
			}
		} else {
			fmt.Println("Error getting renderer from addLiveChatTickerItemAction!")
			fmt.Println(toJSON(item))
		}
	} else if action.Exist("updateLiveChatPollAction") {
		liveChatPoll(action.Get("updateLiveChatPollAction").Get("pollToUpdate").Get("pollRenderer"))
	} else if action.Exist("showLiveChatActionPanelAction") {
		liveChatPoll(action.Get("showLiveChatActionPanelAction").Get("panelToShow").Get("liveChatActionPanelRenderer").Get("contents").Get("pollRenderer"))
	} else if action.Exist("addBannerToLiveChatCommand") { // 釘選
		item := action.Get("addBannerToLiveChatCommand").Get("bannerRenderer").Get("liveChatBannerRenderer").Get("contents")

		if item.Exist("liveChatTextMessageRenderer") {
			rendererProcessor(item.Get("liveChatTextMessageRenderer"), "PinnedTextMessage", video, channel)
		} else if item.Exist("liveChatBannerChatSummaryRenderer") {
		} else {
			fmt.Println("Error getting renderer from addBannerToLiveChatCommand!")
			fmt.Println(toJSON(item))
		}
	} else if action.Exist("removeBannerForLiveChatCommand") { // 取消釘選
	} else if action.Exist("liveChatReportModerationStateCommand") {
	} else if action.Exist("removeChatItemByAuthorAction") {
	} else if action.Exist("removeChatItemAction") {
	} else if action.Exist("closeLiveChatActionPanelAction") {
	} else if action.Exist("replaceChatItemAction") {
	} else {
		fmt.Println("Error getting action!")
		fmt.Println(toJSON(action))
	}
}

func rendererProcessor(renderer *tools.Json, form string, video youtube.Video, channel youtube.Channel) {
	authorChannelId := renderer.Get("authorExternalChannelId").String()
	if !isWatch(authorChannelId) {
		return
	}

	id := renderer.Get("id").String()
	if id == "" || isContain(messageIdList, id) {
		return
	}
	messageIdList = append(messageIdList, id)

	if form == "GiftSend" {
		if renderer.Exist("showItemEndpoint") {
			renderer = renderer.Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer")
		}

		renderer = renderer.Get("header").Get("liveChatSponsorshipsHeaderRenderer")
	} else if form == "Membership" && renderer.Exist("headerPrimaryText") {
		form = "Milestone"
	}

	message := Message{
		Id:       id,
		VideoId:  video.Id,
		AuthorId: authorChannelId,
		Time:     tools.Time(time.Unix(0, int64(renderer.Get("timestampUsec").Int()*1000))),
		Badge:    getBadge(renderer),
		Amount:   renderer.Get("purchaseAmountText").Get("simpleText").String(),
		Text:     getMessage(renderer),
	}

	var template string
	authorChannelName := renderer.Get("authorName").Get("simpleText").String()
	authorChannelUrl := fmt.Sprintf("https://www.youtube.com/channel/%s", authorChannelId)

	switch form {
	case "TextMessage":
		template = fmt.Sprintf("**[%s](<%s>)** 在 **[%s](<%s>)** 的聊天室中留言： `%s`",
			authorChannelName, authorChannelUrl, channel.Title, video.Url, message.Text)
	case "PaidMessage":
		template = fmt.Sprintf("**[%s](<%s>)** 在 **[%s](<%s>)** 的聊天室中購買超級留言(%s)： `%s`",
			authorChannelName, authorChannelUrl, channel.Title, video.Url, message.Amount, message.Text)
	case "PaidSticker":
		template = fmt.Sprintf("**[%s](<%s>)** 在 **[%s](<%s>)** 的聊天室中購買超級貼圖： `%s`",
			authorChannelName, authorChannelUrl, channel.Title, video.Url, message.Text)
	case "Membership":
		template = fmt.Sprintf("**[%s](<%s>)** 成為了 **[%s](<%s>)** 的頻道會員(%s)！",
			authorChannelName, authorChannelUrl, channel.Title, video.Url, message.Badge)
	case "Milestone":
		template = fmt.Sprintf("**[%s](<%s>)** 在 **[%s](<%s>)** 的聊天室中使用里程碑紀念發言： `%s`",
			authorChannelName, authorChannelUrl, channel.Title, video.Url, message.Text)
	case "GiftSend":
		template = fmt.Sprintf("**[%s](<%s>)** 贈送了%s份 **[%s](<%s>)** 的頻道會員！",
			authorChannelName, authorChannelUrl, strings.Split(message.Text, " ")[1], channel.Title, video.Url)
	case "GiftReceive":
		template = fmt.Sprintf("**[%s](<%s>)** 收到了 **[%s](<%s>)** 的頻道會員！",
			authorChannelName, authorChannelUrl, channel.Title, video.Url)
	case "PinnedTextMessage":
		template = fmt.Sprintf("**[%s](<%s>)** 已釘選了 **[%s](<%s>)** 的一則訊息： `%s`",
			channel.Title, video.Url, authorChannelName, authorChannelUrl, message.Text)
	}

	s.ChannelMessageSend(testChannelId, template)
	db.Insert("Message", message.Map())
}

func liveChatPoll(renderer *tools.Json) {
	id := renderer.Get("liveChatPollId").String()
	if id == "" || isContain(messageIdList, id) {
		return
	}
	messageIdList = append(messageIdList, id)

	text := fmt.Sprintf("已發起投票: %s", parseRun(renderer.Get("header").Get("pollHeaderRenderer").Get("pollQuestion")))
	for _, choice := range renderer.Get("choices").JsonArray() {
		text += fmt.Sprintf("\n- %s", parseRun(choice.Get("text")))
	}

	s.ChannelMessageSend(testChannelId, text)
}

func liveChatSetting(renderer *tools.Json) {
	id := renderer.Get("id").String()
	if id == "" || isContain(messageIdList, id) {
		return
	}
	messageIdList = append(messageIdList, id)

	s.ChannelMessageSend(testChannelId, getMessage(renderer))
}

func getBadge(renderer *tools.Json) string {
	var badge string

	for _, badgeData := range renderer.Get("authorBadges").JsonArray() {
		badge += badgeData.Get("liveChatAuthorBadgeRenderer").Get("tooltip").String() + " "
	}

	return badge
}

func getMessage(renderer *tools.Json) string {
	if renderer.Exist("message") {
		var text string

		for _, run := range renderer.Get("message").Get("runs").JsonArray() {
			if run.Exist("text") {
				text += run.Get("text").String()
			} else if run.Exist("emoji") {
				text += run.Get("emoji").Get("shortcuts").Index(0).String()
			} else {
				fmt.Println(toJSON(run))
			}
		}

		if renderer.Exist("subtext") {
			text += "\n" + parseRun(renderer.Get("subtext"))
		}

		return text
	}

	if renderer.Exist("sticker") {
		return renderer.Get("sticker").Get("accessibility").Get("accessibilityData").Get("label").String()
	}

	if renderer.Exist("headerPrimaryText") {
		return parseRun(renderer.Get("headerPrimaryText"))
	}

	if renderer.Exist("headerSubtext") {
		return parseRun(renderer.Get("headerSubtext"))
	}

	if renderer.Exist("primaryText") {
		return parseRun(renderer.Get("primaryText"))
	}

	return ""
}

func parseRun(renderer *tools.Json) string {
	if renderer.Exist("simpleText") {
		return renderer.Get("simpleText").String()
	}

	var text string

	for _, run := range renderer.Get("runs").JsonArray() {
		if run.Exist("text") {
			text += run.Get("text").String()
		} else {
			fmt.Println(toJSON(run))
		}
	}

	return text
}

func isWatch(channelId string) bool {
	for key := range youtube.WatchList {
		if key == channelId {
			return true
		}
	}

	return false
}

func toJSON(item *tools.Json) string {
	jsonBytes, _ := json.Marshal(item)
	return string(jsonBytes)
}

func (message Message) Map() map[string]any {
	messageMap := map[string]any{
		"Id":       message.Id,
		"VideoId":  message.VideoId,
		"AuthorId": message.AuthorId,
		"Time":     message.Time.String(),
		"Text":     message.Text,
	}

	if message.Badge != "" {
		messageMap["Badge"] = message.Badge
	} else {
		messageMap["Badge"] = nil
	}

	if message.Amount != "" {
		messageMap["Amount"] = message.Amount
	} else {
		messageMap["Amount"] = nil
	}

	return messageMap
}
