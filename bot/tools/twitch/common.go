package twitch

import (
	"GoBot/tools"
)

type User struct {
	Id             string
	LoginId        string
	Url            string
	Title          string
	Description    string
	Icon           string
	Thumbnail      string
	ChannelTitle   string
	Color          string
	EmoteMode      bool
	SubscriberMode bool
	UniqueMode     bool
	FollowMode     bool
	FollowTime     int
	SlowMode       bool
	SlowTime       int
	Live           bool
}

type Video struct {
	Id          string
	StreamId    string
	Url         string
	Title       string
	Description string
	Type        string
	Length      tools.Duration
	Created     tools.Time
	Published   tools.Time
}

type Schedule struct {
	UserId        string
	ScheduledTime tools.Time
}

type Badge struct {
	Id          string
	UserId      string
	SetId       string
	Title       string
	Description string
	Image       string
}

type Stamp struct {
	Id     string
	UserId string
	Title  string
	Tier   string
	Type   string
	Format string
	Image  string
}

type ZipBadge struct {
	Old *Badge
	New *Badge
}

type ZipStamp struct {
	Old *Stamp
	New *Stamp
}

func (user User) Map() map[string]any {
	userMap := map[string]any{
		"Id":             user.Id,
		"LoginId":        user.LoginId,
		"Title":          user.Title,
		"Description":    user.Description,
		"ChannelTitle":   user.ChannelTitle,
		"Color":          user.Color,
		"EmoteMode":      user.EmoteMode,
		"SubscriberMode": user.SubscriberMode,
		"UniqueMode":     user.UniqueMode,
		"FollowMode":     user.FollowMode,
		"FollowTime":     user.FollowTime,
		"SlowMode":       user.SlowMode,
		"SlowTime":       user.SlowTime,
		"Live":           user.Live,
	}

	return userMap
}

func (badge Badge) Map() map[string]any {
	badgeMap := map[string]any{
		"Id":          badge.Id,
		"UserId":      badge.UserId,
		"SetId":       badge.SetId,
		"Title":       badge.Title,
		"Description": badge.Description,
	}

	return badgeMap
}

func (stamp Stamp) Map() map[string]any {
	stampMap := map[string]any{
		"Id":     stamp.Id,
		"UserId": stamp.UserId,
		"Title":  stamp.Title,
		"Tier":   stamp.Tier,
		"Type":   stamp.Type,
		"Format": stamp.Format,
	}

	return stampMap
}

func GroupBadge(old, new []Badge) []ZipBadge {
	result := []ZipBadge{}

	badgeMap := map[string]bool{}
	oldMap, newMap := map[string]Badge{}, map[string]Badge{}

	for _, badge := range old {
		badgeMap[badge.Id] = true
		oldMap[badge.Id] = badge
	}

	for _, badge := range new {
		badgeMap[badge.Id] = true
		newMap[badge.Id] = badge
	}

	for badgeId := range badgeMap {
		oldBadge, ok1 := oldMap[badgeId]
		newBadge, ok2 := newMap[badgeId]

		if ok1 && ok2 {
			result = append(result, ZipBadge{Old: &oldBadge, New: &newBadge})
		} else if ok1 && !ok2 {
			result = append(result, ZipBadge{Old: &oldBadge, New: nil})
		} else if !ok1 && ok2 {
			result = append(result, ZipBadge{Old: nil, New: &newBadge})
		}
	}

	return result
}

func GroupStamp(old, new []Stamp) []ZipStamp {
	result := []ZipStamp{}

	stampMap := map[string]bool{}
	oldMap, newMap := map[string]Stamp{}, map[string]Stamp{}

	for _, stamp := range old {
		stampMap[stamp.Id] = true
		oldMap[stamp.Id] = stamp
	}

	for _, stamp := range new {
		stampMap[stamp.Id] = true
		newMap[stamp.Id] = stamp
	}

	for stampId := range stampMap {
		oldStamp, ok1 := oldMap[stampId]
		newStamp, ok2 := newMap[stampId]

		if ok1 && ok2 {
			result = append(result, ZipStamp{Old: &oldStamp, New: &newStamp})
		} else if ok1 && !ok2 {
			result = append(result, ZipStamp{Old: &oldStamp, New: nil})
		} else if !ok1 && ok2 {
			result = append(result, ZipStamp{Old: nil, New: &newStamp})
		}
	}

	return result
}
