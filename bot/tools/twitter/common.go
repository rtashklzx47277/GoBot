package twitter

import "GoBot/tools"

type User struct {
	Id             string
	Username       string
	Name           string
	Url            string
	Description    string
	Location       string
	Link           string
	Latest         string
	Pinned         string
	Icon           string
	Banner         string
	Protected      bool
	Verified       bool
	FollowersCount int
	FollowingCount int
	LikeCount      int
}

type Post struct {
	Id           string
	AuthorId     string
	Url          string
	CreatedTime  tools.Time
	Text         string
	IsRetweeted  bool
	IsReplied    bool
	IsQuoted     bool
	Media        []Media
	ReferencedId string
	EditedId     string
	PollId       string
	Options      []string
}

type Media struct {
	Id   string
	Type string
	Url  string
}

func (user User) Map() map[string]any {
	userMap := map[string]any{
		"Id":             user.Id,
		"Username":       user.Username,
		"Name":           user.Name,
		"Description":    user.Description,
		"Location":       user.Location,
		"Link":           user.Link,
		"Latest":         user.Latest,
		"Pinned":         user.Pinned,
		"Protected":      user.Protected,
		"Verified":       user.Verified,
		"FollowersCount": user.FollowersCount,
		"FollowingCount": user.FollowingCount,
		"LikeCount":      user.LikeCount,
	}

	return userMap
}

func (post Post) Map() map[string]any {
	postMap := map[string]any{
		"Id":           post.Id,
		"AuthorId":     post.AuthorId,
		"CreatedTime":  post.CreatedTime.String(),
		"Text":         post.Text,
		"IsRetweeted":  post.IsRetweeted,
		"IsReplied":    post.IsReplied,
		"IsQuoted":     post.IsQuoted,
		"ReferencedId": post.ReferencedId,
		"EditedId":     post.EditedId,
		"PollId":       post.PollId,
	}

	return postMap
}
