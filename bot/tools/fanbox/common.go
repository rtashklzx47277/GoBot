package fanbox

import "GoBot/tools"

type User struct {
	Id          string
	Url         string
	CreatorId   string
	Name        string
	Description string
	Category    string
	Icon        string
	Banner      string
	Links       []string
	Items       []Item
}

type Item struct {
	Id     string
	UserId string
	Type   string
	Media  string
}

type Plan struct {
	Id          string
	UserId      string
	Title       string
	Fee         int
	Description string
	Image       string
}

type Post struct {
	Id            string
	UserId        string
	Url           string
	Title         string
	Fee           int
	Image         string
	PublishedTime tools.Time
	UpdatedTime   tools.Time
}

type ZipItem struct {
	Old *Item
	New *Item
}

type ZipPlan struct {
	Old *Plan
	New *Plan
}

type ZipPost struct {
	Old *Post
	New *Post
}

func (user User) Map() map[string]any {
	userMap := map[string]any{
		"Id":          user.Id,
		"CreatorId":   user.CreatorId,
		"Name":        user.Name,
		"Description": user.Description,
		"Category":    user.Category,
	}

	return userMap
}

func (item Item) Map() map[string]any {
	itemMap := map[string]any{
		"Id":     item.Id,
		"UserId": item.UserId,
		"Type":   item.Type,
		"Media":  item.Media,
	}

	return itemMap
}

func (paln Plan) Map() map[string]any {
	palnMap := map[string]any{
		"Id":          paln.Id,
		"UserId":      paln.UserId,
		"Title":       paln.Title,
		"Fee":         paln.Fee,
		"Description": paln.Description,
	}

	return palnMap
}

func (post Post) Map() map[string]any {
	postMap := map[string]any{
		"Id":            post.Id,
		"UserId":        post.UserId,
		"Title":         post.Title,
		"Fee":           post.Fee,
		"PublishedTime": post.PublishedTime.String(),
		"UpdatedTime":   post.UpdatedTime.String(),
	}

	return postMap
}

func GroupItem(old, new []Item) []ZipItem {
	result := []ZipItem{}

	itemMap := map[string]bool{}
	oldMap, newMap := map[string]Item{}, map[string]Item{}

	for _, item := range old {
		itemMap[item.Id] = true
		oldMap[item.Id] = item
	}

	for _, item := range new {
		itemMap[item.Id] = true
		newMap[item.Id] = item
	}

	for itemId := range itemMap {
		oldItem, ok1 := oldMap[itemId]
		newItem, ok2 := newMap[itemId]

		if ok1 && ok2 {
			result = append(result, ZipItem{Old: &oldItem, New: &newItem})
		} else if ok1 && !ok2 {
			result = append(result, ZipItem{Old: &oldItem, New: nil})
		} else if !ok1 && ok2 {
			result = append(result, ZipItem{Old: nil, New: &newItem})
		}
	}

	return result
}

func GroupPlan(old, new []Plan) []ZipPlan {
	result := []ZipPlan{}

	planMap := map[string]bool{}
	oldMap, newMap := map[string]Plan{}, map[string]Plan{}

	for _, plan := range old {
		planMap[plan.Id] = true
		oldMap[plan.Id] = plan
	}

	for _, plan := range new {
		planMap[plan.Id] = true
		newMap[plan.Id] = plan
	}

	for planId := range planMap {
		oldPlan, ok1 := oldMap[planId]
		newPlan, ok2 := newMap[planId]

		if ok1 && ok2 {
			result = append(result, ZipPlan{Old: &oldPlan, New: &newPlan})
		} else if ok1 && !ok2 {
			result = append(result, ZipPlan{Old: &oldPlan, New: nil})
		} else if !ok1 && ok2 {
			result = append(result, ZipPlan{Old: nil, New: &newPlan})
		}
	}

	return result
}

func GroupPost(old, new []Post) []ZipPost {
	result := []ZipPost{}

	postMap := map[string]bool{}
	oldMap, newMap := map[string]Post{}, map[string]Post{}

	for _, post := range old {
		postMap[post.Id] = true
		oldMap[post.Id] = post
	}

	for _, post := range new {
		postMap[post.Id] = true
		newMap[post.Id] = post
	}

	for postId := range postMap {
		oldPost, ok1 := oldMap[postId]
		newPost, ok2 := newMap[postId]

		if ok1 && ok2 {
			result = append(result, ZipPost{Old: &oldPost, New: &newPost})
		} else if ok1 && !ok2 {
			result = append(result, ZipPost{Old: &oldPost, New: nil})
		} else if !ok1 && ok2 {
			result = append(result, ZipPost{Old: nil, New: &newPost})
		}
	}

	return result
}
