package main

import (
	"GoBot/tools"
	"GoBot/tools/youtube"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
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
	var apiKey, continuation string
	var err error
	check := true

	for {
		apiKey, continuation, err = getParameters(videoId)
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

		time.Sleep(30 * time.Second)
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
			s.ChannelMessageSend(testChannelId, fmt.Sprintf("**[%s](<%s>)** 聊天室已關閉或直播已轉為會員限定模式！", removeEmoji(db.FindVideoTitle(videoId)), fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoId)))
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
			getMessageData(action, videoId, discordChannelId)
		}

		count = 0
	}
}

func getParameters(videoId string) (string, string, error) {
	url := fmt.Sprintf("https://www.youtube.com/live_chat?v=%s", videoId)
	reader, err := tools.Get(url).AddHeader("User-Agent", tools.UserAgent).Do()
	if err != nil {
		return "", "", err
	}

	data, err := tools.ToString(reader)
	if err != nil {
		return "", "", err
	}

	if strings.Contains(data, "這部直播影片的聊天室已停用") {
		return "", "", ErrorNoChat
	}

	apiKey, ok := tools.Regexp(data, `"INNERTUBE_API_KEY":"([A-z0-9-]*)`)
	if !ok {
		return "", "", fmt.Errorf("failed to get apiKey!\n%w", err)
	}

	continuation, ok := tools.Regexp(data, `"continuation":"([A-z0-9-%]*)`)
	if !ok {
		return "", "", fmt.Errorf("failed to get continuation!\n%w", err)
	}

	return apiKey, continuation, nil
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

func getMessageData(action *tools.Json, videoId, discordChannelId string) {
	if action.Exist("replayChatItemAction") {
		action = action.Get("replayChatItemAction").Get("actions").Index(0)
	}

	if action.Exist("addChatItemAction") {
		item := action.Get("addChatItemAction").Get("item")

		if item.Exist("liveChatTextMessageRenderer") {
			rendererProcessor(item.Get("liveChatTextMessageRenderer"), "TextMessage", videoId, discordChannelId)
		} else if item.Exist("liveChatPaidMessageRenderer") {
			rendererProcessor(item.Get("liveChatPaidMessageRenderer"), "PaidMessage", videoId, discordChannelId)
		} else if item.Exist("liveChatPaidStickerRenderer") {
			rendererProcessor(item.Get("liveChatPaidStickerRenderer"), "PaidSticker", videoId, discordChannelId)
		} else if item.Exist("liveChatMembershipItemRenderer") {
			rendererProcessor(item.Get("liveChatMembershipItemRenderer"), "Membership", videoId, discordChannelId)
		} else if item.Exist("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer") {
			rendererProcessor(item.Get("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer"), "GiftSend", videoId, discordChannelId)
		} else if item.Exist("liveChatSponsorshipsGiftRedemptionAnnouncementRenderer") {
			rendererProcessor(item.Get("liveChatSponsorshipsGiftRedemptionAnnouncementRenderer"), "GiftReceive", videoId, discordChannelId)
		} else if item.Exist("liveChatModeChangeMessageRenderer") {
			liveChatSetting(item.Get("liveChatModeChangeMessageRenderer"))
		} else if item.Exist("liveChatViewerEngagementMessageRenderer") {
		} else if item.Exist("liveChatPlaceholderItemRenderer") {
		} else {
			fmt.Printf("Error getting renderer from addChatItemAction!\n%s\n", youtube.ToJSON(action))
		}
	} else if action.Exist("addLiveChatTickerItemAction") {
		item := action.Get("addLiveChatTickerItemAction").Get("item")

		if item.Exist("liveChatTickerPaidMessageItemRenderer") {
			rendererProcessor(item.Get("liveChatTickerPaidMessageItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatPaidMessageRenderer"), "PaidMessage", videoId, discordChannelId)
		} else if item.Exist("liveChatTickerPaidStickerItemRenderer") {
			rendererProcessor(item.Get("liveChatTickerPaidStickerItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatPaidStickerRenderer"), "PaidSticker", videoId, discordChannelId)
		} else if item.Exist("liveChatTickerSponsorItemRenderer") {
			if item.Get("liveChatTickerSponsorItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Exist("liveChatMembershipItemRenderer") {
				rendererProcessor(item.Get("liveChatTickerSponsorItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatMembershipItemRenderer"), "Membership", videoId, discordChannelId)
			} else if item.Get("liveChatTickerSponsorItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Exist("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer") {
				rendererProcessor(item.Get("liveChatTickerSponsorItemRenderer"), "GiftSend", videoId, discordChannelId)
			} else {
				fmt.Printf("Error getting renderer from liveChatTickerSponsorItemRenderer!\n%s\n", youtube.ToJSON(item))
			}
		} else {
			fmt.Printf("Error getting renderer from addLiveChatTickerItemAction!\n%s\n", youtube.ToJSON(item))
		}
	} else if action.Exist("updateLiveChatPollAction") {
		liveChatPoll(action.Get("updateLiveChatPollAction").Get("pollToUpdate").Get("pollRenderer"), videoId, discordChannelId)
	} else if action.Exist("showLiveChatActionPanelAction") {
		liveChatPoll(action.Get("showLiveChatActionPanelAction").Get("panelToShow").Get("liveChatActionPanelRenderer").Get("contents").Get("pollRenderer"), videoId, discordChannelId)
	} else if action.Exist("addBannerToLiveChatCommand") { // 釘選
		item := action.Get("addBannerToLiveChatCommand").Get("bannerRenderer").Get("liveChatBannerRenderer").Get("contents")

		if item.Exist("liveChatTextMessageRenderer") {
		} else if item.Exist("liveChatBannerChatSummaryRenderer") {
		} else {
			fmt.Printf("Error getting renderer from addBannerToLiveChatCommand!\n%s\n", youtube.ToJSON(item))
		}
	} else if action.Exist("removeBannerForLiveChatCommand") { // 取消釘選
	} else if action.Exist("liveChatReportModerationStateCommand") {
	} else if action.Exist("removeChatItemByAuthorAction") {
	} else if action.Exist("removeChatItemAction") {
	} else if action.Exist("closeLiveChatActionPanelAction") {
	} else if action.Exist("replaceChatItemAction") {
	} else {
		fmt.Printf("Error getting action!\n%s\n", youtube.ToJSON(action))
	}
}

func rendererProcessor(renderer *tools.Json, form, videoId, discordChannelId string) {
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
	videoTitle := removeEmoji(db.FindVideoTitle(videoId))
	videoUrl := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoId)

	switch form {
	case "TextMessage":
		template = fmt.Sprintf("： `%s`", message.Text)
	case "PaidMessage":
		template = fmt.Sprintf(" 發送了超級留言(%s)： `%s`", message.Amount, message.Text)
	case "PaidSticker":
		template = fmt.Sprintf(" 發送了超級貼圖： `%s`", message.Text)
	case "Milestone":
		template = fmt.Sprintf(" 發送了里程碑紀念留言： `%s`", message.Text)
	case "Membership":
		template = fmt.Sprintf(" 成為了頻道會員(%s)！", message.Badge)
	case "GiftSend":
		template = fmt.Sprintf(" 贈送了%s份頻道會員！", strings.Split(message.Text, " ")[1])
	case "GiftReceive":
		template = " 收到了的頻道會員贈禮！"
	}

	s.ChannelMessageSend(discordChannelId, fmt.Sprintf("**[%s](<%s>)**%s [%s](<%s>)", authorName, authorUrl, template, videoTitle, videoUrl))
	db.Insert("Message", message.Map())
}

func liveChatPoll(renderer *tools.Json, videoId, discordChannelId string) {
	id := renderer.Get("liveChatPollId").String()
	if id == "" || tools.IsContain(messageIdList, id) {
		return
	}
	messageIdList = append(messageIdList, id)

	videoTitle := removeEmoji(db.FindVideoTitle(videoId))
	videoUrl := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoId)

	text := fmt.Sprintf("已發起投票: %s [%s](<%s>)", youtube.ParseRun(renderer.Get("header").Get("pollHeaderRenderer").Get("pollQuestion")), videoTitle, videoUrl)
	for _, choice := range renderer.Get("choices").JsonArray() {
		text += fmt.Sprintf("\n- %s", youtube.ParseRun(choice.Get("text")))
	}

	s.ChannelMessageSend(discordChannelId, text)
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

	if badge == "" {
		return badge
	}

	return badge[:len(badge)-1]
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

func removeEmoji(input string) string {
	return regexp.MustCompile(`[\x{1F300}-\x{1F5FF}]|[\x{1F600}-\x{1F64F}]|[\x{1F680}-\x{1F6FF}]|[\x{1F900}-\x{1F9FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]`).ReplaceAllString(input, "")
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
