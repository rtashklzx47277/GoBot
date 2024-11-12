package youtube

import (
	"GoBot/tools"
	"fmt"
	"os"
	"strings"
)

var (
	secure_3PSID      = os.Getenv("SECURE_3PSID")
	secure_3PSIDTS    = os.Getenv("SECURE_3PSIDTS")
	ErrorNoMembership = fmt.Errorf("membership is not yet available")
)

func GetMemberShip(channelId string) ([]Badge, []Stamp, []Perk, error) {
	url := fmt.Sprintf("https://www.youtube.com/channel/%s/membership", channelId)
	reader, err := tools.Get(url).AddCookie("__Secure-3PSID", secure_3PSID).AddCookie("__Secure-3PSIDTS", secure_3PSIDTS).Do()
	if err != nil {
		return []Badge{}, []Stamp{}, []Perk{}, err
	}

	data, err := tools.ToString(reader)
	if err != nil {
		return []Badge{}, []Stamp{}, []Perk{}, err
	}

	match, ok := tools.Regexp(data, `ytInitialData = (.+?);\s*<\/script>`)
	if !ok {
		return []Badge{}, []Stamp{}, []Perk{}, fmt.Errorf("failed to get ytInitialData!\n%w", err)
	}

	jsonData, err := tools.StringToJson(match)
	if err != nil {
		return []Badge{}, []Stamp{}, []Perk{}, err
	}

	tab, ok := getTab(jsonData.Get("contents").Get("twoColumnBrowseResultsRenderer").Get("tabs"), "Membership", "會員資格")
	if !ok {
		return []Badge{}, []Stamp{}, []Perk{}, ErrorNoMembership
	}

	var badges []Badge
	var stamps []Stamp
	var perks []Perk

	for _, item := range tab.Get("tabRenderer").Get("content").Get("sectionListRenderer").Get("contents").Index(0).Get("sponsorshipsExpandablePerksRenderer").Get("expandableItems").JsonArray() {
		renderer := item.Get("sponsorshipsPerkRenderer")

		if renderer.Exist("loyaltyBadges") {
			for _, item := range renderer.Get("loyaltyBadges").Get("sponsorshipsLoyaltyBadgesRenderer").Get("loyaltyBadges").JsonArray() {
				badge := Badge{
					Label: parseLabel(parseRun(item.Get("sponsorshipsLoyaltyBadgeRenderer").Get("title"))),
					Image: getThumbnail(item.Get("sponsorshipsLoyaltyBadgeRenderer").Get("icon")),
				}

				badge.Author.Id = channelId

				badges = append(badges, badge)
			}
		} else if renderer.Exist("images") {
			for _, item := range renderer.Get("images").JsonArray() {
				stamp := Stamp{
					Label: item.Get("accessibility").Get("accessibilityData").Get("label").String(),
					Image: getThumbnail(item),
				}

				stamp.Author.Id = channelId

				stamps = append(stamps, stamp)
			}
		} else {
			perk := Perk{
				Title:       parseRun(renderer.Get("title")),
				Description: parseRun(renderer.Get("description")),
			}

			perk.Author.Id = channelId

			perks = append(perks, perk)
		}
	}

	return badges, stamps, perks, nil
}

func parseLabel(label string) string {
	if strings.HasPrefix(label, "New") || strings.HasPrefix(label, "新") {
		return "0"
	}

	return strings.Split(label, " ")[1]
}
