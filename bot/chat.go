package main

import (
	"GoBot/tools"
	"GoBot/tools/youtube"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

type Message struct {
	Id        string
	VideoId   string
	Type      string
	ChannelId string
	Time      tools.Time
	Badge     string
	Amount    string
	Text      string
}

var ErrorNoChat = fmt.Errorf("chat is not yet available")

func LiveChat(videoId, discordChannelId string) {
	var channelTitle, apiKey, continuation string
	var err error
	check := true

	for {
		channelTitle, apiKey, continuation, err = getParameters(videoId)
		if err == nil {
			break
		}

		if errors.Is(err, ErrorNoChat) {
			fmt.Printf("直播影片的聊天室已停用，重新嘗試讀取聊天室 (%s)\n", videoId)

			if db.CheckVideoStatus(videoId) == 0 {
				check = false
				break
			}
		} else if db.CheckVideoMember(videoId) {
			return
		} else {
			s.ChannelMessageSend(testChannelId, fmt.Sprintf("failed to get \"%s\" parameters: %v\n", videoId, err))
		}

		time.Sleep(300 * time.Second)
	}

	if !check {
		return
	}

	messageIdList = append(messageIdList, db.Distinct("message", videoId)...)

	fmt.Printf("Start Getting Video Chat: %v\n", videoId)
	defer fmt.Printf("Stop Getting Video Chat: %v\n", videoId)

	count := 0

	for {
		if count == 5 {
			s.ChannelMessageSend(testChannelId, fmt.Sprintf("**[%s](<%s>)** 聊天室已關閉或直播已轉為會員限定模式！", db.FindVideoTitle(videoId), fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoId)))
			break
		}

		data, err := getChatData(apiKey, continuation)
		if err != nil {
			fmt.Printf("failed to get \"%s\" chat data: %v\n", videoId, err)
			count++
			continue
		}

		if !data.Exist("continuationContents") {
			count++
			continue
		}

		continuations := data.Get("continuationContents").Get("liveChatContinuation").Get("continuations").Index(0)

		if continuations.Exist("timedContinuationData") {
			continuation = continuations.Get("timedContinuationData").Get("continuation").String()
		} else if continuations.Exist("invalidationContinuationData") {
			continuation = continuations.Get("invalidationContinuationData").Get("continuation").String()
		}

		for _, action := range data.Get("continuationContents").Get("liveChatContinuation").Get("actions").JsonArray() {
			getMessageData(action, videoId, channelTitle, discordChannelId)
		}

		count = 0
	}
}

func getParameters(videoId string) (string, string, string, error) {
	url := fmt.Sprintf("https://www.youtube.com/live_chat?v=%s", videoId)
	reader, err := tools.Get(url).AddHeader("User-Agent", tools.UserAgent).Do()
	if err != nil {
		return "", "", "", err
	}

	data, err := tools.ToString(reader)
	if err != nil {
		return "", "", "", err
	}

	if strings.Contains(data, "這部直播影片的聊天室已停用") {
		return "", "", "", ErrorNoChat
	}

	channelTitle, ok := tools.Regexp(data, `{"authorName":{"simpleText":"(.+?)"}`)
	if !ok {
		return "", "", "", fmt.Errorf("failed to get channel title!\n%w", err)
	}

	apiKey, ok := tools.Regexp(data, `"INNERTUBE_API_KEY":"([A-z0-9-]*)`)
	if !ok {
		return "", "", "", fmt.Errorf("failed to get apiKey!\n%w", err)
	}

	continuation, ok := tools.Regexp(data, `"continuation":"([A-z0-9-%]*)`)
	if !ok {
		return "", "", "", fmt.Errorf("failed to get continuation!\n%w", err)
	}

	return channelTitle, apiKey, continuation, nil
}

func getChatData(apiKey, continuation string) (*tools.Json, error) {
	url := fmt.Sprintf("https://www.youtube.com/youtubei/v1/live_chat/get_live_chat?key=%s", apiKey)
	payload, err := getPayload(continuation)
	if err != nil {
		return &tools.Json{}, err
	}

	reader, err := tools.Post(url, payload).AddHeader("User-Agent", tools.UserAgent).Do()
	if err != nil {
		return &tools.Json{}, err
	}

	data, err := tools.ToJson(reader)
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
				"clientVersion": tools.ClientVersion,
			},
		},
		"continuation": continuation,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create payload!\n%w", err)
	}

	return io.NopCloser(bytes.NewReader(payloadBytes)), nil
}

func getMessageData(action *tools.Json, videoId, channelTitle, discordChannelId string) {
	if action.Exist("addChatItemAction") {
		item := action.Get("addChatItemAction").Get("item")

		if item.Exist("liveChatTextMessageRenderer") {
			rendererProcessor(item.Get("liveChatTextMessageRenderer"), "TextMessage", videoId, channelTitle, discordChannelId)
		} else if item.Exist("liveChatPaidMessageRenderer") {
			rendererProcessor(item.Get("liveChatPaidMessageRenderer"), "PaidMessage", videoId, channelTitle, discordChannelId)
		} else if item.Exist("liveChatPaidStickerRenderer") {
			rendererProcessor(item.Get("liveChatPaidStickerRenderer"), "PaidSticker", videoId, channelTitle, discordChannelId)
		} else if item.Exist("liveChatMembershipItemRenderer") {
			rendererProcessor(item.Get("liveChatMembershipItemRenderer"), "Membership", videoId, channelTitle, discordChannelId)
		} else if item.Exist("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer") {
			rendererProcessor(item.Get("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer"), "GiftSend", videoId, channelTitle, discordChannelId)
		} else if item.Exist("liveChatSponsorshipsGiftRedemptionAnnouncementRenderer") {
			rendererProcessor(item.Get("liveChatSponsorshipsGiftRedemptionAnnouncementRenderer"), "GiftReceive", videoId, channelTitle, discordChannelId)
		} else if item.Exist("liveChatModeChangeMessageRenderer") {
			liveChatSetting(item.Get("liveChatModeChangeMessageRenderer"))
		} else if item.Exist("liveChatViewerEngagementMessageRenderer") {
		} else if item.Exist("liveChatPlaceholderItemRenderer") {
		} else {
			fmt.Println("Error getting renderer from addChatItemAction!")
			fmt.Println(youtube.ToJSON(item))
		}
	} else if action.Exist("addLiveChatTickerItemAction") {
		item := action.Get("addLiveChatTickerItemAction").Get("item")

		if item.Exist("liveChatTickerPaidMessageItemRenderer") {
			rendererProcessor(item.Get("liveChatTickerPaidMessageItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatPaidMessageRenderer"), "PaidMessage", videoId, channelTitle, discordChannelId)
		} else if item.Exist("liveChatTickerPaidStickerItemRenderer") {
			rendererProcessor(item.Get("liveChatTickerPaidStickerItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatPaidStickerRenderer"), "PaidSticker", videoId, channelTitle, discordChannelId)
		} else if item.Exist("liveChatTickerSponsorItemRenderer") {
			if item.Get("liveChatTickerSponsorItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Exist("liveChatMembershipItemRenderer") {
				rendererProcessor(item.Get("liveChatTickerSponsorItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatMembershipItemRenderer"), "Membership", videoId, channelTitle, discordChannelId)
			} else if item.Get("liveChatTickerSponsorItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Exist("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer") {
				rendererProcessor(item.Get("liveChatTickerSponsorItemRenderer"), "GiftSend", videoId, channelTitle, discordChannelId)
			} else {
				fmt.Println("Error getting renderer from liveChatTickerSponsorItemRenderer!")
				fmt.Println(youtube.ToJSON(item))
			}
		} else {
			fmt.Println("Error getting renderer from addLiveChatTickerItemAction!")
			fmt.Println(youtube.ToJSON(item))
		}
	} else if action.Exist("updateLiveChatPollAction") {
		liveChatPoll(action.Get("updateLiveChatPollAction").Get("pollToUpdate").Get("pollRenderer"))
	} else if action.Exist("showLiveChatActionPanelAction") {
		liveChatPoll(action.Get("showLiveChatActionPanelAction").Get("panelToShow").Get("liveChatActionPanelRenderer").Get("contents").Get("pollRenderer"))
	} else if action.Exist("addBannerToLiveChatCommand") { // 釘選
		item := action.Get("addBannerToLiveChatCommand").Get("bannerRenderer").Get("liveChatBannerRenderer").Get("contents")

		if item.Exist("liveChatTextMessageRenderer") {
		} else if item.Exist("liveChatBannerChatSummaryRenderer") {
		} else {
			fmt.Println("Error getting renderer from addBannerToLiveChatCommand!")
			fmt.Println(youtube.ToJSON(item))
		}
	} else if action.Exist("removeBannerForLiveChatCommand") { // 取消釘選
	} else if action.Exist("liveChatReportModerationStateCommand") {
	} else if action.Exist("removeChatItemByAuthorAction") {
	} else if action.Exist("removeChatItemAction") {
	} else if action.Exist("closeLiveChatActionPanelAction") {
	} else if action.Exist("replaceChatItemAction") {
	} else {
		fmt.Println("Error getting action!")
		fmt.Println(youtube.ToJSON(action))
	}
}

func rendererProcessor(renderer *tools.Json, form, videoId, channelTitle, discordChannelId string) {
	authorId := renderer.Get("authorExternalChannelId").String()
	messageId := renderer.Get("id").String()

	if form == "GiftSend" {
		if renderer.Exist("showItemEndpoint") {
			renderer = renderer.Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer")
		}

		renderer = renderer.Get("header").Get("liveChatSponsorshipsHeaderRenderer")
	} else if form == "Membership" && renderer.Exist("headerPrimaryText") {
		form = "Milestone"
	}

	badge := getBadge(renderer)

	if !check(authorId, badge) || messageId == "" || tools.IsContain(messageIdList, messageId) {
		return
	}

	messageIdList = append(messageIdList, messageId)

	message := Message{
		Id:        messageId,
		VideoId:   videoId,
		Type:      form,
		ChannelId: authorId,
		Time:      tools.Time(time.Unix(0, int64(renderer.Get("timestampUsec").Int()*1000))),
		Badge:     badge,
		Amount:    renderer.Get("purchaseAmountText").Get("simpleText").String(),
		Text:      getMessage(renderer),
	}

	var template string
	authorName := renderer.Get("authorName").Get("simpleText").String()
	authorUrl := fmt.Sprintf("https://www.youtube.com/channel/%s", authorId)
	videoUrl := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoId)

	switch form {
	case "TextMessage":
		template = fmt.Sprintf("**[%s](<%s>)** 在 **[%s](<%s>)** 的聊天室中留言： `%s`",
			authorName, authorUrl, channelTitle, videoUrl, message.Text)
	case "PaidMessage":
		template = fmt.Sprintf("**[%s](<%s>)** 在 **[%s](<%s>)** 的聊天室中購買超級留言(%s)： `%s`",
			authorName, authorUrl, channelTitle, videoUrl, message.Amount, message.Text)
	case "PaidSticker":
		template = fmt.Sprintf("**[%s](<%s>)** 在 **[%s](<%s>)** 的聊天室中購買超級貼圖： `%s`",
			authorName, authorUrl, channelTitle, videoUrl, message.Text)
	case "Membership":
		template = fmt.Sprintf("**[%s](<%s>)** 成為了 **[%s](<%s>)** 的頻道會員(%s)！",
			authorName, authorUrl, channelTitle, videoUrl, message.Badge)
	case "Milestone":
		template = fmt.Sprintf("**[%s](<%s>)** 在 **[%s](<%s>)** 的聊天室中使用里程碑紀念發言： `%s`",
			authorName, authorUrl, channelTitle, videoUrl, message.Text)
	case "GiftSend":
		template = fmt.Sprintf("**[%s](<%s>)** 贈送了%s份 **[%s](<%s>)** 的頻道會員！",
			authorName, authorUrl, strings.Split(message.Text, " ")[1], channelTitle, videoUrl)
	case "GiftReceive":
		template = fmt.Sprintf("**[%s](<%s>)** 收到了 **[%s](<%s>)** 的頻道會員！",
			authorName, authorUrl, channelTitle, videoUrl)
	}

	s.ChannelMessageSend(discordChannelId, template)
	db.Insert("Message", message.Map())
}

func liveChatPoll(renderer *tools.Json) {
	id := renderer.Get("liveChatPollId").String()
	if id == "" || tools.IsContain(messageIdList, id) {
		return
	}
	messageIdList = append(messageIdList, id)

	text := fmt.Sprintf("已發起投票: %s", youtube.ParseRun(renderer.Get("header").Get("pollHeaderRenderer").Get("pollQuestion")))
	for _, choice := range renderer.Get("choices").JsonArray() {
		text += fmt.Sprintf("\n- %s", youtube.ParseRun(choice.Get("text")))
	}

	s.ChannelMessageSend(testChannelId, text)
}

func liveChatSetting(renderer *tools.Json) {
	id := renderer.Get("id").String()
	if id == "" || tools.IsContain(messageIdList, id) {
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
				fmt.Println(youtube.ToJSON(run))
			}
		}

		if renderer.Exist("subtext") {
			text += "\n" + youtube.ParseRun(renderer.Get("subtext"))
		}

		return text
	}

	if renderer.Exist("sticker") {
		return renderer.Get("sticker").Get("accessibility").Get("accessibilityData").Get("label").String()
	}

	if renderer.Exist("headerPrimaryText") {
		return youtube.ParseRun(renderer.Get("headerPrimaryText"))
	}

	if renderer.Exist("headerSubtext") {
		return youtube.ParseRun(renderer.Get("headerSubtext"))
	}

	if renderer.Exist("primaryText") {
		return youtube.ParseRun(renderer.Get("primaryText"))
	}

	return ""
}

func check(channelId, badge string) bool {
	if strings.Contains(badge, "Owner") || strings.Contains(badge, "Moderator") {
		return true
	}

	for key := range tools.ChannelList {
		if key == channelId {
			return true
		}
	}

	return false
}

func (message Message) Map() map[string]any {
	messageMap := map[string]any{
		"Id":        message.Id,
		"VideoId":   message.VideoId,
		"Type":      message.Type,
		"ChannelId": message.ChannelId,
		"Time":      message.Time.String(),
		"Text":      message.Text,
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
