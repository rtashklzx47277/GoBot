package youtube

import (
	"GoBot/tools"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	messageIdList = []string{}
	clientVersion = "2.20240122.00.00"
)

type Message struct {
	Id        string
	Badge     string
	Amount    string
	Time      tools.Time
	Onwer     bool
	Moderator bool
	Verified  bool
	Member    bool
	Author    Channel
}

func LiveChat(videoId string) {
	apiKey, continuation, err := getParameters(videoId)
	if err != nil {
		return
	}

	for {
		data, err := getChatData(apiKey, continuation)
		if err != nil {
			continue
		}

		if !data.Exist("continuationContents") {
			fmt.Println("cant find continuationContents")
			fmt.Println(toJSON(data))
			continue
		}

		continuations := data.Get("continuationContents").Get("liveChatContinuation").Get("continuations").Index(0)

		if continuations.Exist("timedContinuationData") {
			continuation = continuations.Get("timedContinuationData").Get("continuation").String()
		} else if continuations.Exist("invalidationContinuationData") {
			continuation = continuations.Get("invalidationContinuationData").Get("continuation").String()
		}

		for _, action := range data.Get("continuationContents").Get("liveChatContinuation").Get("actions").JsonArray() {
			getMessageData(action)
		}
	}
}

func getParameters(videoId string) (string, string, error) {
	url := fmt.Sprintf("https://www.youtube.com/live_chat?v=%s", videoId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Set("User-Agent", userAgent)

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

func getMessageData(action *tools.Json) {
	if action.Exist("addChatItemAction") {
		item := action.Get("addChatItemAction").Get("item")

		if item.Exist("liveChatTextMessageRenderer") {
			renderer := item.Get("liveChatTextMessageRenderer")
			liveChatTextMessage(renderer)
		} else if item.Exist("liveChatPaidMessageRenderer") {
			renderer := item.Get("liveChatPaidMessageRenderer")
			liveChatPaidMessage(renderer)
		} else if item.Exist("liveChatPaidStickerRenderer") {
			renderer := item.Get("liveChatPaidStickerRenderer")
			liveChatPaidSticker(renderer)
		} else if item.Exist("liveChatMembershipItemRenderer") {
			renderer := item.Get("liveChatMembershipItemRenderer")
			liveChatMembership(renderer)
		} else if item.Exist("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer") {
			renderer := item.Get("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer")
			liveChatGiftSend(renderer)
		} else if item.Exist("liveChatSponsorshipsGiftRedemptionAnnouncementRenderer") {
			renderer := item.Get("liveChatSponsorshipsGiftRedemptionAnnouncementRenderer")
			liveChatGiftReceive(renderer)
		} else if item.Exist("liveChatModeChangeMessageRenderer") {
			renderer := item.Get("liveChatModeChangeMessageRenderer")
			liveChatSetting(renderer)
		} else if item.Exist("liveChatViewerEngagementMessageRenderer") {
		} else if item.Exist("liveChatPlaceholderItemRenderer") {
		} else {
			fmt.Println("Error getting renderer from addChatItemAction")
			fmt.Println(toJSON(item))
			return
		}
	} else if action.Exist("addLiveChatTickerItemAction") {
		item := action.Get("addLiveChatTickerItemAction").Get("item")

		if item.Exist("liveChatTickerPaidMessageItemRenderer") {
			renderer := item.Get("liveChatTickerPaidMessageItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatPaidMessageRenderer")
			liveChatPaidMessage(renderer)
		} else if item.Exist("liveChatTickerPaidStickerItemRenderer") {
			renderer := item.Get("liveChatTickerPaidStickerItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatPaidStickerRenderer")
			liveChatPaidSticker(renderer)
		} else if item.Exist("liveChatMembershipItemRenderer") {
			fmt.Println("Error getting renderer from liveChatMembershipItemRenderer")
			fmt.Println(toJSON(item))
			renderer := item.Get("liveChatMembershipItemRenderer")
			liveChatMembership(renderer)
		} else if item.Exist("liveChatTickerSponsorItemRenderer") {
			renderer := item.Get("liveChatTickerSponsorItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatMembershipItemRenderer")
			liveChatMembership(renderer)
		} else {
			fmt.Println("Error getting renderer from addLiveChatTickerItemAction")
			fmt.Println(toJSON(item))
			return
		}
	} else if action.Exist("addBannerToLiveChatCommand") { // 釘選
		item := action.Get("addBannerToLiveChatCommand").Get("bannerRenderer").Get("liveChatBannerRenderer").Get("contents")

		if item.Exist("liveChatTextMessageRenderer") {
			renderer := item.Get("liveChatTextMessageRenderer")
			liveChatTextMessage(renderer)
		} else {
			fmt.Println("Error getting renderer from addBannerToLiveChatCommand")
			fmt.Println(toJSON(item))
			return
		}
	} else if action.Exist("removeBannerForLiveChatCommand") { // 取消釘選
	} else if action.Exist("liveChatReportModerationStateCommand") {
	} else if action.Exist("removeChatItemByAuthorAction") {
	} else if action.Exist("removeChatItemAction") {
	} else if action.Exist("replaceChatItemAction") {
	} else {
		fmt.Println("Error getting action")
		fmt.Println(toJSON(action))
		return
	}
}

func liveChatTextMessage(renderer *tools.Json) {
	if !renderer.Exist("id") || check(renderer) {
		return
	}

	timestamp := getTimestamp(renderer)
	authorChannelId := renderer.Get("authorExternalChannelId").String()
	authorName := renderer.Get("authorName").Get("simpleText").String()
	badge := getBadge(renderer)
	message := getMessage(renderer)

	if !strings.Contains(badge, "Owner") {
		return
	}

	fmt.Println(timestamp)
	fmt.Println(authorChannelId)
	fmt.Println(authorName, badge)
	fmt.Println(message)
	fmt.Println("==================================================")
}

func liveChatPaidMessage(renderer *tools.Json) {
	if !renderer.Exist("id") || check(renderer) {
		return
	}

	timestamp := getTimestamp(renderer)
	authorChannelId := renderer.Get("authorExternalChannelId").String()
	authorName := renderer.Get("authorName").Get("simpleText").String()
	badge := getBadge(renderer)
	purchaseAmountText := renderer.Get("purchaseAmountText").Get("simpleText").String()
	message := getMessage(renderer)

	fmt.Println(timestamp)
	fmt.Println(authorChannelId)
	fmt.Println(authorName, badge)
	fmt.Println(purchaseAmountText)

	if message != "" {
		fmt.Println(message)
	}

	fmt.Println("==================================================")
}

func liveChatPaidSticker(renderer *tools.Json) {
	if !renderer.Exist("id") || check(renderer) {
		return
	}

	timestamp := getTimestamp(renderer)
	authorChannelId := renderer.Get("authorExternalChannelId").String()
	authorName := renderer.Get("authorName").Get("simpleText").String()
	badge := getBadge(renderer)
	purchaseAmountText := renderer.Get("purchaseAmountText").Get("simpleText").String()
	message := renderer.Get("sticker").Get("accessibility").Get("accessibilityData").Get("label").String()

	fmt.Println(timestamp)
	fmt.Println(authorChannelId)
	fmt.Println(authorName, badge)
	fmt.Println(purchaseAmountText)
	fmt.Println(message)
	fmt.Println("==================================================")
}

func liveChatMembership(renderer *tools.Json) {
	if !renderer.Exist("id") || check(renderer) {
		return
	}

	timestamp := getTimestamp(renderer)
	authorChannelId := renderer.Get("authorExternalChannelId").String()
	authorName := renderer.Get("authorName").Get("simpleText").String()
	badge := getBadge(renderer)
	message := getMessage(renderer)

	fmt.Println(timestamp)
	fmt.Println(authorChannelId)
	fmt.Println(authorName, badge)
	fmt.Println(message)
	fmt.Println("==================================================")
}

func liveChatGiftSend(renderer *tools.Json) {
	if !renderer.Exist("id") || check(renderer) {
		return
	}

	timestamp := getTimestamp(renderer)
	authorChannelId := renderer.Get("authorExternalChannelId").String()
	authorName := renderer.Get("header").Get("liveChatSponsorshipsHeaderRenderer").Get("authorName").Get("simpleText").String()
	badge := getBadge(renderer.Get("header").Get("liveChatSponsorshipsHeaderRenderer"))
	message := getMessage(renderer.Get("header").Get("liveChatSponsorshipsHeaderRenderer"))

	fmt.Println(timestamp)
	fmt.Println(authorChannelId)
	fmt.Println(authorName, badge)
	fmt.Println(message)
	fmt.Println("==================================================")
}

func liveChatGiftReceive(renderer *tools.Json) {
	if !renderer.Exist("id") || check(renderer) {
		return
	}

	timestamp := getTimestamp(renderer)
	authorChannelId := renderer.Get("authorExternalChannelId").String()
	authorName := renderer.Get("authorName").Get("simpleText").String()
	badge := getBadge(renderer)
	message := getMessage(renderer)

	fmt.Println(timestamp)
	fmt.Println(authorChannelId)
	fmt.Println(authorName, badge)
	fmt.Println(message)
	fmt.Println("==================================================")
}

func liveChatSetting(renderer *tools.Json) {
	if !renderer.Exist("id") || check(renderer) {
		return
	}

	timestamp := getTimestamp(renderer)
	message := getMessage(renderer)

	fmt.Println(timestamp)
	fmt.Println(message)
	fmt.Println("==================================================")
}

func getTimestamp(renderer *tools.Json) string {
	return time.Unix(0, int64(renderer.Get("timestampUsec").Int()*1000)).Format("2006/01/02 15:04:05")
}

func getBadge(renderer *tools.Json) string {
	var badge string

	for _, badgeData := range renderer.Get("authorBadges").JsonArray() {
		badge += badgeData.Get("liveChatAuthorBadgeRenderer").Get("tooltip").String() + " "
	}

	return badge
}

func getMessage(renderer *tools.Json) string {
	if renderer.Exist("headerSubtext") {
		return parseRun(renderer, "headerSubtext")
	}

	if renderer.Exist("primaryText") {
		return parseRun(renderer, "primaryText")
	}

	var text, subtext string

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
		subtext += "\n"

		for _, run := range renderer.Get("subtext").Get("runs").JsonArray() {
			if run.Exist("text") {
				text += run.Get("text").String()
			} else {
				fmt.Println(toJSON(run))
			}
		}
	}

	return text + subtext
}

func parseRun(renderer *tools.Json, class string) string {
	var text string

	for _, run := range renderer.Get(class).Get("runs").JsonArray() {
		if run.Exist("text") {
			text += run.Get("text").String()
		} else {
			fmt.Println(toJSON(run))
		}
	}

	return text
}

func check(renderer *tools.Json) bool {
	id := renderer.Get("id").String()

	if tools.IsContain(messageIdList, id) {
		return true
	}

	messageIdList = tools.Append(messageIdList, id)

	return false
}

func toJSON(item *tools.Json) string {
	jsonBytes, _ := json.Marshal(item)
	return string(jsonBytes)
}
