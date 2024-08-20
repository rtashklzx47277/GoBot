package youtube

import (
	"GoBot/tools"
	"fmt"
)

func GetMemberShip(channelId string) error {
	url := fmt.Sprintf("https://www.youtube.com/channel/%s/membership", channelId)
	// reader, err := tools.Get(url).AddCookie("__Secure-3PSID", secure_3PSID).AddCookie("__Secure-3PSIDTS", secure_3PSIDTS).Do()
	reader, err := tools.Get(url).AddCookie("__Secure-3PSID", "g.a000lAjRYRLuNcdPQBKY5hemSclkrs3amB189wAIvY09Gwtd-EfwtIcP4ncwwKNfmyX5LmpYqAACgYKAdESARASFQHGX2Mi-0MgW0nXwYj3jIKtAVc7jBoVAUF8yKo9fxZOniGPS3ONf8cna0S70076").AddCookie("__Secure-3PSIDTS", "sidts-CjIB3EgAEjOwkR6BC-WNEE9-ICaw2w7Ik4GVybukHHtzeEjxN7lZqPBn-_gqO1E8WuLL0BAA").Do()
	if err != nil {
		return err
	}

	data, err := tools.ToString(reader)
	if err != nil {
		return err
	}

	match, ok := tools.Regexp(data, `ytInitialData = (.+?);\s*<\/script>`)
	if !ok {
		return fmt.Errorf("failed to get ytInitialData!\n%w", err)
	}

	jsonData, err := tools.StringToJson(match)
	if err != nil {
		return err
	}

	for _, item := range getTab(jsonData.Get("contents").Get("twoColumnBrowseResultsRenderer").Get("tabs")).Get("tabRenderer").Get("content").Get("sectionListRenderer").Get("contents").Index(0).Get("sponsorshipsExpandablePerksRenderer").Get("expandableItems").JsonArray() {
		renderer := item.Get("sponsorshipsPerkRenderer")

		if renderer.Exist("loyaltyBadges") {
			for _, badge := range renderer.Get("loyaltyBadges").Get("sponsorshipsLoyaltyBadgesRenderer").Get("loyaltyBadges").JsonArray() {
				icon := getThumbnail(badge.Get("sponsorshipsLoyaltyBadgeRenderer").Get("icon"))
				title := parseRun(badge.Get("sponsorshipsLoyaltyBadgeRenderer").Get("title"))

				fmt.Println(title)
				fmt.Println(icon)
			}
		} else if renderer.Exist("images") {
			for _, stamp := range renderer.Get("images").JsonArray() {
				thumbnail := getThumbnail(stamp)
				label := stamp.Get("accessibility").Get("accessibilityData").Get("label")

				fmt.Println(label)
				fmt.Println(thumbnail)
			}
		} else {
			title := parseRun(renderer.Get("title"))
			description := parseRun(renderer.Get("description"))

			fmt.Println(title)
			fmt.Println(description)
		}
	}

	return nil
}
