package twitcasting

type User struct {
	Id          string
	ScreenId    string
	Url         string
	Title       string
	Description string
	Icon        string
	Live        bool
}

type Stream struct {
	Id        string
	Url       string
	Title     string
	Subtitle  string
	Thumbnail string
	IsRecord  bool
}

func (user User) Map() map[string]any {
	userMap := map[string]any{
		"Id":          user.Id,
		"ScreenId":    user.ScreenId,
		"Title":       user.Title,
		"Description": user.Description,
		"Live":        user.Live,
	}

	return userMap
}
