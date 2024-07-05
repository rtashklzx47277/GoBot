package main

import (
	"GoBot/tools"
	"fmt"
	"time"
)

// Owner Verified
// New member
// Member (1 month)
// Member (4 years)

var messageIds = []string{}

func LiveChatbyOriginal(videoId string) {
	apiKey, continuation, err := getParameters(videoId)
	if err != nil {
		return
	}

	count := 0

	for {
		if count == 5 {
			s.ChannelMessageSend(testChannelId, "聊天室已關閉或直播已轉為會員限定模式！")
			break
		}

		data, err := getChatData(apiKey, continuation)
		if err != nil {
			continue
		}

		if !data.Exist("continuationContents") {
			count++
			fmt.Println("Can't find continuationContents!")
			continue
		}

		count = 0
		continuations := data.Get("continuationContents").Get("liveChatContinuation").Get("continuations").Index(0)

		if continuations.Exist("timedContinuationData") {
			continuation = continuations.Get("timedContinuationData").Get("continuation").String()
		} else if continuations.Exist("invalidationContinuationData") {
			continuation = continuations.Get("invalidationContinuationData").Get("continuation").String()
		}

		for _, action := range data.Get("continuationContents").Get("liveChatContinuation").Get("actions").JsonArray() {
			getMessageDataOriginal(action)
		}
	}
}

func getMessageDataOriginal(action *tools.Json) {
	if action.Exist("addChatItemAction") {
		item := action.Get("addChatItemAction").Get("item")

		if item.Exist("liveChatTextMessageRenderer") { // message
			// rendererProcessorOriginal(item.Get("liveChatTextMessageRenderer"), "TextMessage")
		} else if item.Exist("liveChatPaidMessageRenderer") { // message
			rendererProcessorOriginal(item.Get("liveChatPaidMessageRenderer"), "PaidMessage")
		} else if item.Exist("liveChatPaidStickerRenderer") { // sticker
			rendererProcessorOriginal(item.Get("liveChatPaidStickerRenderer"), "PaidSticker")
		} else if item.Exist("liveChatMembershipItemRenderer") { // headerPrimaryText headerSubtext (message)
			rendererProcessorOriginal(item.Get("liveChatMembershipItemRenderer"), "Membership")
		} else if item.Exist("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer") { // primaryText
			rendererProcessorOriginal(item.Get("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer"), "GiftSend")
		} else if item.Exist("liveChatSponsorshipsGiftRedemptionAnnouncementRenderer") { // message
			rendererProcessorOriginal(item.Get("liveChatSponsorshipsGiftRedemptionAnnouncementRenderer"), "GiftReceive")
		} else if item.Exist("liveChatModeChangeMessageRenderer") {
			printMessageType(item.Get("liveChatModeChangeMessageRenderer"), "addChatItemAction -> liveChatModeChangeMessageRenderer")
			liveChatSettingOriginal(item.Get("liveChatModeChangeMessageRenderer"))
		} else if item.Exist("liveChatViewerEngagementMessageRenderer") {
		} else if item.Exist("liveChatPlaceholderItemRenderer") {
		} else {
			fmt.Println("Error getting renderer from addChatItemAction!")
			fmt.Println(toJSON(item))
		}
	} else if action.Exist("addLiveChatTickerItemAction") {
		item := action.Get("addLiveChatTickerItemAction").Get("item")

		if item.Exist("liveChatTickerPaidMessageItemRenderer") { // message
			rendererProcessorOriginal(item.Get("liveChatTickerPaidMessageItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatPaidMessageRenderer"), "PaidMessage")
		} else if item.Exist("liveChatTickerPaidStickerItemRenderer") { // sticker
			rendererProcessorOriginal(item.Get("liveChatTickerPaidStickerItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatPaidStickerRenderer"), "PaidSticker")
		} else if item.Exist("liveChatTickerSponsorItemRenderer") {
			if item.Get("liveChatTickerSponsorItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Exist("liveChatMembershipItemRenderer") { // headerPrimaryText headerSubtext (message)
				rendererProcessorOriginal(item.Get("liveChatTickerSponsorItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatMembershipItemRenderer"), "Membership")
			} else if item.Get("liveChatTickerSponsorItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Exist("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer") { // primaryText
				rendererProcessorOriginal(item.Get("liveChatTickerSponsorItemRenderer"), "GiftSend")
			} else {
				fmt.Println("Error getting renderer from liveChatTickerSponsorItemRenderer!")
				fmt.Println(toJSON(item))
			}
		} else {
			fmt.Println("Error getting renderer from addLiveChatTickerItemAction!")
			fmt.Println(toJSON(item))
		}
	} else if action.Exist("updateLiveChatPollAction") {
		liveChatPollOriginal(action.Get("updateLiveChatPollAction").Get("pollToUpdate").Get("pollRenderer"))
	} else if action.Exist("showLiveChatActionPanelAction") {
		liveChatPollOriginal(action.Get("showLiveChatActionPanelAction").Get("panelToShow").Get("liveChatActionPanelRenderer").Get("contents").Get("pollRenderer"))
	} else if action.Exist("addBannerToLiveChatCommand") { // 釘選
		item := action.Get("addBannerToLiveChatCommand").Get("bannerRenderer").Get("liveChatBannerRenderer").Get("contents")

		if item.Exist("liveChatTextMessageRenderer") {
			rendererProcessorOriginal(item.Get("liveChatTextMessageRenderer"), "PinnedTextMessage")
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

func rendererProcessorOriginal(renderer *tools.Json, form string) {
	messageId := renderer.Get("id").String()
	if messageId == "" || isContain(messageIds, messageId) {
		return
	}

	messageIds = append(messageIds, messageId)

	authorChannelId := renderer.Get("authorExternalChannelId").String()

	if form == "GiftSend" {
		if renderer.Exist("showItemEndpoint") {
			renderer = renderer.Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer")
		}

		renderer = renderer.Get("header").Get("liveChatSponsorshipsHeaderRenderer")
	} else if form == "Membership" && renderer.Exist("headerPrimaryText") {
		form = "Milestone"
	} else if form == "PinnedTextMessage" {
		fmt.Println("已釘選訊息！")
	}

	authorName := renderer.Get("authorName").Get("simpleText").String()
	time := tools.Time(time.Unix(0, int64(renderer.Get("timestampUsec").Int()*1000))).String()
	badge := getBadge(renderer)
	amount := renderer.Get("purchaseAmountText").Get("simpleText").String()
	text := getMessage(renderer)

	if authorName == "" {
		fmt.Println(toJSON(renderer))
	}

	fmt.Println(form)
	fmt.Printf("%s(%s) %s %s\n", authorName, authorChannelId, badge, amount)

	if form != "GiftSend" {
		fmt.Println(time)
	}

	if text != "" {
		fmt.Println(text)
	}

	fmt.Println("===========================================================================")
}

func liveChatSettingOriginal(renderer *tools.Json) {
	if !renderer.Exist("id") {
		return
	}

	id := renderer.Get("id").String()
	if isContain(messageIds, id) {
		return
	}

	messageIds = append(messageIds, id)

	fmt.Println(getMessage(renderer))
}

func liveChatPollOriginal(renderer *tools.Json) {
	if !renderer.Exist("liveChatPollId") {
		return
	}

	id := renderer.Get("liveChatPollId").String()
	if isContain(messageIds, id) {
		return
	}

	messageIds = append(messageIds, id)

	fmt.Printf("已發起投票: %s\n", parseRun(renderer.Get("header").Get("pollHeaderRenderer").Get("pollQuestion")))
	for _, choice := range renderer.Get("choices").JsonArray() {
		fmt.Println(parseRun(choice.Get("text")))
	}
}

func printMessageType(renderer *tools.Json, form string) {
	fmt.Println(form)
	fmt.Println("message:", renderer.Exist("message"))
	fmt.Println("sticker:", renderer.Exist("sticker"))
	fmt.Println("headerPrimaryText:", renderer.Exist("headerPrimaryText"))
	fmt.Println("headerSubtext:", renderer.Exist("headerSubtext"))
	fmt.Println("primaryText:", renderer.Exist("primaryText"))
}
