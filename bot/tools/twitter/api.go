package twitter

import (
	"GoBot/tools"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	bearerTokenList = map[int]string{
		1: os.Getenv("TWITTER_BEARER_TOKEN_1"),
		2: os.Getenv("TWITTER_BEARER_TOKEN_2"),
		3: os.Getenv("TWITTER_BEARER_TOKEN_3"),
		4: os.Getenv("TWITTER_BEARER_TOKEN_4"),
		5: os.Getenv("TWITTER_BEARER_TOKEN_5"),
		6: os.Getenv("TWITTER_BEARER_TOKEN_6"),
		7: os.Getenv("TWITTER_BEARER_TOKEN_7"),
		8: os.Getenv("TWITTER_BEARER_TOKEN_8"),
	}
	mediaMap = map[string]Media{}
	pollMap  = map[string][]string{}
	counter1 = 0
	counter2 = 0
)

func getData(path string, queries ...string) (*tools.Json, error) {
	var count *int

	if strings.Contains(path, "tweets") {
		count = &counter1
	} else {
		count = &counter2
	}

	for range 8 {
		req := tools.Get(path).AddHeader("Authorization", fmt.Sprintf("Bearer %s", bearerTokenList[*count%8+1]))

		for i := 0; i < len(queries); i += 2 {
			req = req.AddQuery(queries[i], queries[i+1])
		}

		reader, err := req.Do()
		if errors.Is(err, tools.ErrorTooManyRequests) {
			*count++
			continue
		} else if err != nil {
			return &tools.Json{}, err
		}

		data, err := tools.ToJson(reader)
		if err != nil {
			return &tools.Json{}, err
		}

		return data, nil
	}

	return &tools.Json{}, tools.ErrorTooManyRequests
}

func GetUser(username string) (User, error) {
	url := fmt.Sprintf("https://api.twitter.com/2/users/by/username/%s", username)
	data, err := getData(url, "user.fields", "id,name,username,description,entities,location,most_recent_tweet_id,pinned_tweet_id,profile_banner_url,profile_image_url,protected,public_metrics,verified")
	if errors.Is(err, tools.ErrorTooManyRequests) {
		return User{}, tools.ErrorTooManyRequests
	} else if err != nil {
		return User{}, fmt.Errorf("failed to get user data!\n%w", err)
	}

	item := data.Get("data")
	user := User{
		Id:             item.Get("id").String(),
		Username:       item.Get("username").String(),
		Name:           item.Get("name").String(),
		Link:           item.Get("entities").Get("url").Get("urls").Index(0).Get("expanded_url").String(),
		Description:    item.Get("description").String(),
		Location:       item.Get("location").String(),
		Latest:         item.Get("most_recent_tweet_id").String(),
		Pinned:         item.Get("pinned_tweet_id").String(),
		Icon:           item.Get("profile_image_url").Replace("_normal", "", 1),
		Banner:         item.Get("profile_banner_url").String(),
		Protected:      item.Get("protected").Bool(),
		Verified:       item.Get("verified").Bool(),
		FollowersCount: item.Get("public_metrics").Get("followers_count").Int(),
		FollowingCount: item.Get("public_metrics").Get("following_count").Int(),
		LikeCount:      item.Get("public_metrics").Get("like_count").Int(),
	}

	user.Url = fmt.Sprintf("https://x.com/%s", user.Username)

	return user, nil
}

func GetUsers(usernames ...string) (map[string]User, error) {
	users := map[string]User{}

	url := "https://api.twitter.com/2/users/by"
	data, err := getData(url,
		"usernames", strings.Join(usernames, ","),
		"user.fields", "id,name,username,description,entities,location,most_recent_tweet_id,pinned_tweet_id,profile_banner_url,profile_image_url,protected,public_metrics,verified")
	if errors.Is(err, tools.ErrorTooManyRequests) {
		return map[string]User{}, tools.ErrorTooManyRequests
	} else if err != nil {
		return map[string]User{}, fmt.Errorf("failed to get user data!\n%w", err)
	}

	for _, item := range data.Get("data").JsonArray() {
		user := User{
			Id:             item.Get("id").String(),
			Username:       item.Get("username").String(),
			Name:           item.Get("name").String(),
			Url:            item.Get("entities").Get("url").Get("urls").Index(0).Get("expanded_url").String(),
			Description:    item.Get("description").String(),
			Location:       item.Get("location").String(),
			Pinned:         item.Get("pinned_tweet_id").String(),
			Icon:           item.Get("profile_image_url").Replace("_normal", "", 1),
			Banner:         item.Get("profile_banner_url").String(),
			Protected:      item.Get("protected").Bool(),
			Verified:       item.Get("verified").Bool(),
			FollowersCount: item.Get("public_metrics").Get("followers_count").Int(),
			FollowingCount: item.Get("public_metrics").Get("following_count").Int(),
			LikeCount:      item.Get("public_metrics").Get("like_count").Int(),
		}

		users[user.Username] = user
	}

	return users, nil
}

func GetUsername(userId string) (string, error) {
	url := fmt.Sprintf("https://api.twitter.com/2/users/%s", userId)
	data, err := getData(url, "user.fields", "id,name,username,description,entities,location,most_recent_tweet_id,pinned_tweet_id,profile_banner_url,profile_image_url,protected,public_metrics,verified")
	if errors.Is(err, tools.ErrorTooManyRequests) {
		return "", tools.ErrorTooManyRequests
	} else if err != nil {
		return "", fmt.Errorf("failed to get user data!\n%w", err)
	}

	return data.Get("data").Get("username").String(), nil
}

func GetTimeline(userId, sinceId string) ([]Post, error) {
	var posts []Post

	since, _ := strconv.Atoi(sinceId)
	url := fmt.Sprintf("https://api.twitter.com/2/users/%s/tweets", userId)
	data, err := getData(url,
		"max_results", "100",
		"since_id", strconv.Itoa(since+1),
		"media.fields", "type,url,variants",
		"poll.fields", "id,options,end_datetime",
		"tweet.fields", "id,author_id,created_at,text,entities,referenced_tweets,edit_history_tweet_ids",
		"expansions", "attachments.media_keys,attachments.poll_ids")
	if errors.Is(err, tools.ErrorTooManyRequests) {
		return []Post{}, tools.ErrorTooManyRequests
	} else if err != nil {
		return []Post{}, fmt.Errorf("failed to get user tweets!\n%w", err)
	}

	for _, media := range data.Get("includes").Get("media").JsonArray() {
		mediaId := media.Get("media_key").String()

		if _, ok := mediaMap[mediaId]; !ok {
			var mediaUrl string
			mediaType := media.Get("type").String()

			if mediaType == "photo" {
				mediaUrl = media.Get("url").String()
			} else if mediaType == "video" {
				mediaUrl = media.Get("variants").Index(-2).Get("url").String()
			}

			mediaMap[mediaId] = Media{
				Id:   mediaId,
				Type: mediaType,
				Url:  mediaUrl,
			}
		}
	}

	for _, poll := range data.Get("includes").Get("polls").JsonArray() {
		pollId := poll.Get("id").String()

		if _, ok := pollMap[pollId]; !ok {
			for _, option := range poll.Get("options").JsonArray() {
				pollMap[pollId] = append(pollMap[pollId], option.Get("label").String())
			}
		}
	}

	for _, item := range data.Get("data").JsonArray() {
		post := getPostStruct(item)
		posts = append(posts, post)
	}

	return posts, nil
}

func getPostStruct(data *tools.Json) Post {
	post := Post{
		Id:          data.Get("id").String(),
		AuthorId:    data.Get("author_id").String(),
		CreatedTime: data.Get("created_at").Time(),
	}

	post.Url = fmt.Sprintf("https://x.com/x/status/%s", post.Id)

	if data.Exist("referenced_tweets") {
		referencedType := data.Get("referenced_tweets").Index(0).Get("type").String()
		switch referencedType {
		case "retweeted":
			post.IsRetweeted = true
		case "replied_to":
			post.IsReplied = true
		case "quoted":
			post.IsQuoted = true
		}

		post.ReferencedId = data.Get("referenced_tweets").Index(0).Get("id").String()
	}

	if post.IsRetweeted {
		return post
	}

	post.Text = data.Get("text").String()

	for _, this := range data.Get("entities").Get("urls").JsonArray() {
		if this.Exist("media_key") {
			post.Media = append(post.Media, mediaMap[this.Get("media_key").String()])
		} else if !this.Get("display_url").HasPrefix("x.com/") {
			post.Text = strings.Replace(post.Text, this.Get("url").String(), this.Get("unwound_url").String(), 1)
		}
	}

	ids := data.Get("edit_history_tweet_ids").Array()
	idsLen := len(ids)
	if idsLen > 1 && ids[idsLen-1] != post.Id {
		post.EditedId = ids[idsLen-1].(string)
	}

	if data.Get("attachments").Exist("poll_ids") {
		post.PollId = data.Get("attachments").Get("poll_ids").Index(0).String()
		post.Options = pollMap[post.PollId]
	}

	return post
}
