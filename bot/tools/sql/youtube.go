package sql

import (
	"GoBot/tools/youtube"
	"database/sql"
	"fmt"
)

func (mySQL *MySQL) FindChannel(channelId string) youtube.Channel {
	var channel youtube.Channel
	var customId, title, description sql.NullString
	var subscriberCount, viewCount sql.NullInt64

	query := "SELECT Id, CustomId, Title, Description, SubscriberCount, ViewCount FROM Channel WHERE Id = ?"
	err := mySQL.db.QueryRow(query, channelId).Scan(&channel.Id, &customId, &title, &description, &subscriberCount, &viewCount)
	if err != nil {
		fmt.Println(err)
	}

	channel.CustomId = handleNullString(customId)
	channel.Title = handleNullString(title)
	channel.Description = handleNullString(description)
	channel.SubscriberCount = handleNullInt(subscriberCount)
	channel.ViewCount = handleNullInt(viewCount)

	channel.Icon = fmt.Sprintf("/bot/media/Youtube/%s/Icon/%s.jpg", channel.Id, channel.Id)
	channel.Banner = fmt.Sprintf("/bot/media/Youtube/%s/Banner/%s.jpg", channel.Id, channel.Id)
	channel.Url = fmt.Sprintf("https://www.youtube.com/channel/%s", channel.Id)

	return channel
}

func (mySQL *MySQL) FindVideos(channelId string) []youtube.Video {
	query := "SELECT Id, Title, Description, Length, ViewCount, LiveStatus, PublishedTime, Comment, Private, Music FROM Video WHERE ChannelId = ?"
	rows, err := mySQL.db.Query(query, channelId)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var videos []youtube.Video

	for rows.Next() {
		var video youtube.Video
		var title, description, length, publishedTime sql.NullString
		var viewCount, liveStatus sql.NullInt64

		err := rows.Scan(&video.Id, &title, &description, &length, &viewCount, &liveStatus, &publishedTime, &video.Comment, &video.Private, &video.Music)
		if err != nil {
			fmt.Println(err)
		}

		video.Title = handleNullString(title)
		video.Description = handleNullString(description)
		video.Length = stringToDuration(handleNullString(length))
		video.ViewCount = handleNullInt(viewCount)
		video.LiveStatus = handleNullInt(liveStatus)
		video.PublishedTime = stringToTime(handleNullString(publishedTime))

		video.Thumbnail = fmt.Sprintf("/bot/media/Youtube/%s/Video/%s.jpg", channelId, video.Id)
		video.Url = fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.Id)

		videos = append(videos, video)
	}

	return videos
}

func (mySQL *MySQL) FindLivestreams(channelId string) []youtube.Video {
	query := "SELECT DISTINCT Id, Video.ChannelId, Title, LiveStatus, ScheduledTime FROM Video LEFT JOIN Collab ON Video.Id = Collab.VideoId " +
		"WHERE (Video.ChannelId = ? OR Collab.ChannelId = ?) AND LiveStatus <> ? AND Private = ?"
	rows, err := mySQL.db.Query(query, channelId, channelId, 0, 0)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var livestreams []youtube.Video

	for rows.Next() {
		var livestream youtube.Video
		var title, scheduledTime sql.NullString
		var liveStatus sql.NullInt64

		err := rows.Scan(&livestream.Id, &livestream.Author.Id, &title, &liveStatus, &scheduledTime)
		if err != nil {
			fmt.Println(err)
		}

		livestream.Title = handleNullString(title)
		livestream.LiveStatus = handleNullInt(liveStatus)
		livestream.ScheduledTime = stringToTime(handleNullString(scheduledTime))

		if channelId == livestream.Author.Id {
			livestream.Thumbnail = fmt.Sprintf("/bot/media/Youtube/%s/Video/%s.jpg", channelId, livestream.Id)
		} else {
			livestream.Thumbnail = fmt.Sprintf("/bot/media/Youtube/%s/Collab/%s.jpg", channelId, livestream.Id)
		}

		livestream.Url = fmt.Sprintf("https://www.youtube.com/watch?v=%s", livestream.Id)

		livestream.Author.Url = fmt.Sprintf("https://www.youtube.com/channel/%s", livestream.Author.Id)

		livestreams = append(livestreams, livestream)
	}

	return livestreams
}

func (mySQL *MySQL) FindPlaylists(channelId string) []youtube.Playlist {
	query := "SELECT Id, Title, Description FROM Playlist WHERE channelId = ?"
	rows, err := mySQL.db.Query(query, channelId)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var playlists []youtube.Playlist

	for rows.Next() {
		var playlist youtube.Playlist
		var title, description sql.NullString

		err := rows.Scan(&playlist.Id, &title, &description)
		if err != nil {
			fmt.Println(err)
		}

		playlist.Title = handleNullString(title)
		playlist.Description = handleNullString(description)

		playlist.Thumbnail = fmt.Sprintf("/bot/media/Youtube/%s/Playlist/%s.jpg", channelId, playlist.Id)
		playlist.Url = fmt.Sprintf("https://www.youtube.com/playlist?list=%s", playlist.Id)

		playlists = append(playlists, playlist)
	}

	return playlists
}

func (mySQL *MySQL) FindPlaylistItems(playlistId string) []youtube.Video {
	query := "SELECT Id, ChannelId, Title FROM Video RIGHT JOIN PlaylistItem ON PlaylistItem.VideoId = Video.Id WHERE PlaylistItem.PlaylistId = ?"
	rows, err := mySQL.db.Query(query, playlistId)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var playlistItems []youtube.Video

	for rows.Next() {
		var playlistItem youtube.Video
		var title sql.NullString

		err := rows.Scan(&playlistItem.Id, &playlistItem.Author.Id, &title)
		if err != nil {
			fmt.Println(err)
		}

		playlistItem.Title = handleNullString(title)

		playlistItem.Thumbnail = fmt.Sprintf("/bot/media/Youtube/%s/Video/%s.jpg", playlistItem.Author.Id, playlistItem.Id)
		playlistItem.Url = fmt.Sprintf("https://www.youtube.com/watch?v=%s", playlistItem.Id)

		playlistItems = append(playlistItems, playlistItem)
	}

	return playlistItems
}

func (mySQL *MySQL) FindComments(channelId string) []youtube.Comment {
	query := "SELECT c.Id, c.ParentId, c.Text, c.PublishedTime, c.UpdatedTime, " +
		"c1.Id, c1.Title, v.Id, v.Title, c.Canceled FROM Comment c " +
		"LEFT JOIN Channel c1 ON c1.Id = c.ChannelId " +
		"LEFT JOIN Video v ON v.Id = c.VideoId " +
		"LEFT JOIN Channel c2 ON c2.Id = v.ChannelId " +
		"WHERE c2.Id = ?"
	rows, err := mySQL.db.Query(query, channelId)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var comments []youtube.Comment

	for rows.Next() {
		var comment youtube.Comment
		var parentId, text, publishedTime, updatedTime, authorChannelTitle, videoTitle sql.NullString

		err := rows.Scan(&comment.Id, &parentId, &text, &publishedTime, &updatedTime,
			&comment.Author.Id, &authorChannelTitle, &comment.Video.Id, &videoTitle, &comment.Canceled)
		if err != nil {
			fmt.Println(err)
		}

		comment.ParentId = handleNullString(parentId)
		comment.Text = handleNullString(text)
		comment.PublishedTime = stringToTime(handleNullString(publishedTime))
		comment.UpdatedTime = stringToTime(handleNullString(updatedTime))
		comment.Author.Title = handleNullString(authorChannelTitle)
		comment.Video.Title = handleNullString(videoTitle)

		comment.Author.Url = fmt.Sprintf("https://www.youtube.com/channel/%s", comment.Author.Id)
		comment.Video.Url = fmt.Sprintf("https://www.youtube.com/watch?v=%s", comment.Video.Id)

		comments = append(comments, comment)
	}

	return comments
}

func (mySQL *MySQL) FindMusic(channelId string) []youtube.Video {
	query := "SELECT Id, Title, ViewCount FROM Video LEFT JOIN Collab ON Video.Id = Collab.VideoId WHERE Video.Music = ? AND Video.Private = ? AND (Video.ChannelId = ? OR Collab.ChannelId = ?)"
	rows, err := mySQL.db.Query(query, 1, 0, channelId, channelId)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var videos []youtube.Video

	for rows.Next() {
		var video youtube.Video
		var title sql.NullString
		var viewCount sql.NullInt64

		err := rows.Scan(&video.Id, &title, &viewCount)
		if err != nil {
			fmt.Println(err)
		}

		video.Title = handleNullString(title)
		video.ViewCount = handleNullInt(viewCount)

		video.Thumbnail = fmt.Sprintf("/bot/media/Youtube/%s/Video/%s.jpg", channelId, video.Id)
		video.Url = fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.Id)

		videos = append(videos, video)
	}

	return videos
}

func (mySQL *MySQL) Distinct(target, channelId string) []string {
	var query string
	var values []any

	switch target {
	case "video":
		query = "SELECT DISTINCT Id FROM Video WHERE ChannelId = ?"
		values = append(values, channelId)
	case "public":
		query = "SELECT DISTINCT Id FROM Video WHERE ChannelId = ? AND Member = ?"
		values = append(values, channelId, 0)
	case "member":
		query = "SELECT DISTINCT Id FROM Video WHERE ChannelId = ? AND Member = ?"
		values = append(values, channelId, 1)
	case "livestream":
		query = "SELECT DISTINCT Video.Id FROM Video LEFT JOIN Collab ON Video.Id = Collab.VideoId WHERE (Video.ChannelId = ? OR Collab.ChannelId = ?) AND Video.LiveStatus <> ? AND Video.Private = ?"
		values = append(values, channelId, channelId, 0, 0)
	case "music":
		query = "SELECT DISTINCT Video.Id FROM Video LEFT JOIN Collab ON Video.Id = Collab.VideoId WHERE (Video.ChannelId = ? OR Collab.ChannelId = ?) AND Video.Music = ? AND Video.Private = ?"
		values = append(values, channelId, channelId, 1, 0)
	case "comment":
		query = "SELECT DISTINCT Comment.Id FROM Comment LEFT JOIN Video ON Comment.VideoId = Video.Id WHERE Video.ChannelId = ? AND ParentId IS NULL"
		values = append(values, channelId)
	case "reply":
		query = "SELECT DISTINCT Comment.Id FROM Comment LEFT JOIN Video ON Comment.VideoId = Video.Id WHERE Video.ChannelId = ? AND ParentId IS NOT NULL"
		values = append(values, channelId)
	case "collab":
		query = "SELECT DISTINCT VideoId FROM Collab WHERE ChannelId = ?"
		values = append(values, channelId)
	}

	rows, err := mySQL.db.Query(query, values...)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var result []string

	for rows.Next() {
		var value string

		err := rows.Scan(&value)
		if err != nil {
			fmt.Println(err)
		}
		result = append(result, value)
	}

	return result
}

func (mySQL *MySQL) CompelteComment(comment youtube.Comment) youtube.Comment {
	if comment.ParentId != "" {
		var videoId sql.NullString

		query := "SELECT c2.VideoId FROM Comment c1 LEFT JOIN Comment c2 ON c1.ParentId = c2.Id WHERE c1.Id = ?"

		err := mySQL.db.QueryRow(query, comment.Id).Scan(&videoId)
		if err != nil {
			fmt.Println(err)
		}

		comment.Video.Id = handleNullString(videoId)
		comment.Video.Url = fmt.Sprintf("https://www.youtube.com/watch?v=%s", comment.Video.Id)
	}

	var authorTitle, videoTitle sql.NullString

	query := "SELECT Channel.Title, Video.Title FROM Comment LEFT JOIN Video ON Comment.VideoId = Video.Id LEFT JOIN Channel ON Comment.ChannelId = Channel.Id WHERE Comment.Id = ?"

	err := mySQL.db.QueryRow(query, comment.Id).Scan(&authorTitle, &videoTitle)
	if err != nil {
		fmt.Println(err)
	}

	comment.Author.Title = handleNullString(authorTitle)
	comment.Video.Title = handleNullString(videoTitle)

	return comment
}
