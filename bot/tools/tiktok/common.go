package tiktok

type User struct {
	Id          string
	Url         string
	ShortId     string
	UniqueId    string
	Title       string
	Description string
	Icon        string
	FollowCount int
}

func (user User) Map() map[string]any {
	userMap := map[string]any{
		"Id":          user.Id,
		"ShortId":     user.ShortId,
		"UniqueId":    user.UniqueId,
		"Title":       user.Title,
		"Description": user.Description,
		"FollowCount": user.FollowCount,
	}

	return userMap
}
