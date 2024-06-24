package main

import (
	"GoBot/tools"
	"fmt"
	"time"
)

var messageIds = []string{}

func LiveChatbyOriginal(videoId string) {
	apiKey, continuation, err := getParameters(videoId)
	if err != nil {
		return
	}

	for {
		data, err := getChatData(apiKey, continuation)
		if err != nil {
			continue
		}

		if data.Get("contents").Get("messageRenderer").Get("text").Get("runs").Index(0).Get("text").String() == "Sorry, live chat is currently unavailable." {
			fmt.Println("聊天室已關閉或已轉為會員限定模式！")
			return
		}

		if !data.Exist("continuationContents") {
			fmt.Println("Can't find continuationContents!")
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
			getMessageDataOriginal(action)
		}
	}
}

func getMessageDataOriginal(action *tools.Json) {
	if action.Exist("addChatItemAction") {
		item := action.Get("addChatItemAction").Get("item")

		if item.Exist("liveChatTextMessageRenderer") {
			rendererProcessorOriginal(item.Get("liveChatTextMessageRenderer"), "TextMessage")
		} else if item.Exist("liveChatPaidMessageRenderer") {
			rendererProcessorOriginal(item.Get("liveChatPaidMessageRenderer"), "PaidMessage")
		} else if item.Exist("liveChatPaidStickerRenderer") {
			rendererProcessorOriginal(item.Get("liveChatPaidStickerRenderer"), "PaidSticker")
		} else if item.Exist("liveChatMembershipItemRenderer") {
			rendererProcessorOriginal(item.Get("liveChatMembershipItemRenderer"), "Membership")
		} else if item.Exist("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer") {
			rendererProcessorOriginal(item.Get("liveChatSponsorshipsGiftPurchaseAnnouncementRenderer"), "GiftSend")
		} else if item.Exist("liveChatSponsorshipsGiftRedemptionAnnouncementRenderer") {
			rendererProcessorOriginal(item.Get("liveChatSponsorshipsGiftRedemptionAnnouncementRenderer"), "GiftReceive")
		} else if item.Exist("liveChatModeChangeMessageRenderer") {
			liveChatSettingOriginal(item.Get("liveChatModeChangeMessageRenderer"))
		} else if item.Exist("liveChatViewerEngagementMessageRenderer") {
		} else if item.Exist("liveChatPlaceholderItemRenderer") {
		} else {
			fmt.Println("Error getting renderer from addChatItemAction!")
			fmt.Println(toJSON(item))
			return
		}
	} else if action.Exist("addLiveChatTickerItemAction") {
		item := action.Get("addLiveChatTickerItemAction").Get("item")

		if item.Exist("liveChatTickerPaidMessageItemRenderer") {
			rendererProcessorOriginal(item.Get("liveChatTickerPaidMessageItemRenderer"), "PaidMessage")
		} else if item.Exist("liveChatTickerPaidStickerItemRenderer") {
			rendererProcessorOriginal(item.Get("liveChatTickerPaidStickerItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatPaidStickerRenderer"), "PaidSticker")
		} else if item.Exist("liveChatMembershipItemRenderer") {
			rendererProcessorOriginal(item.Get("liveChatMembershipItemRenderer"), "Membership")
		} else if item.Exist("liveChatTickerSponsorItemRenderer") {
			rendererProcessorOriginal(item.Get("liveChatTickerSponsorItemRenderer").Get("showItemEndpoint").Get("showLiveChatItemEndpoint").Get("renderer").Get("liveChatMembershipItemRenderer"), "Membership")
		} else {
			fmt.Println("Error getting renderer from addLiveChatTickerItemAction!")
			fmt.Println(toJSON(item))
			return
		}
	} else if action.Exist("addBannerToLiveChatCommand") { // 釘選
		item := action.Get("addBannerToLiveChatCommand").Get("bannerRenderer").Get("liveChatBannerRenderer").Get("contents")

		if item.Exist("liveChatTextMessageRenderer") {
			fmt.Println("已釘選訊息！")
			rendererProcessorOriginal(item.Get("liveChatTextMessageRenderer"), "TextMessage")
		} else {
			fmt.Println("Error getting renderer from addBannerToLiveChatCommand!")
			fmt.Println(toJSON(item))
			return
		}
	} else if action.Exist("removeBannerForLiveChatCommand") { // 取消釘選
	} else if action.Exist("liveChatReportModerationStateCommand") {
	} else if action.Exist("removeChatItemByAuthorAction") {
	} else if action.Exist("removeChatItemAction") {
	} else if action.Exist("replaceChatItemAction") {
	} else {
		fmt.Println("Error getting action!")
		fmt.Println(toJSON(action))
		return
	}
}

func rendererProcessorOriginal(renderer *tools.Json, form string) {
	if form == "TextMessage" {
		return
	}

	if !renderer.Exist("id") {
		return
	}

	messageId := renderer.Get("id").String()
	if isContain(messageIds, messageId) {
		return
	}

	messageIds = append(messageIds, messageId)

	authorChannelId := renderer.Get("authorExternalChannelId").String()

	if form == "GiftSend" {
		renderer = renderer.Get("header").Get("liveChatSponsorshipsHeaderRenderer")
	}

	authorName := renderer.Get("authorName").Get("simpleText").String()
	time := tools.Time(time.Unix(0, int64(renderer.Get("timestampUsec").Int()*1000))).String()
	badge := getBadge(renderer)
	amount := renderer.Get("purchaseAmountText").Get("simpleText").String()
	text := getMessage(renderer)

	fmt.Printf("%s(%s) %s %s\n", authorName, authorChannelId, badge, amount)
	fmt.Println(time)
	fmt.Println(text)
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

	s.ChannelMessageSend(testChannelId, getMessage(renderer))
}
