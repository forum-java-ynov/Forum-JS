package backend

import (
	"database/sql"
	"strings"
)

func getFilteredPosts(theme string, userID int, likedOnly bool, hispost bool) ([]Post, error) {
	query := `
       SELECT 
          posts.id,
          posts.title,
          posts.content,
          posts.image_path,
          posts.theme,
          COALESCE(users.full_name, users.username, 'Utilisateur'),
          COUNT(DISTINCT pl_all.id) as likes,
          COUNT(DISTINCT pd_all.id) as dislikes,
          MAX(CASE WHEN pl_all.user_id = ? THEN 1 ELSE 0 END) as user_liked,
          MAX(CASE WHEN pd_all.user_id = ? THEN 1 ELSE 0 END) as user_disliked
       FROM posts
       LEFT JOIN users ON posts.user_id = users.id
       LEFT JOIN post_like pl_all ON posts.id = pl_all.post_id
       LEFT JOIN post_dislike pd_all ON posts.id = pd_all.post_id
    `

	args := []interface{}{userID, userID}
	var conditions []string

	if theme != "" {
		conditions = append(conditions, "posts.theme = ?")
		args = append(args, theme)
	}

	if likedOnly && userID > 0 {
		conditions = append(conditions, "posts.id IN (SELECT post_id FROM post_like WHERE user_id = ?)")
		args = append(args, userID)
	}

	if hispost && userID > 0 {
		conditions = append(conditions, "posts.user_id = ?")
		args = append(args, userID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " GROUP BY posts.id"

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var id, likes, dislikes int
		var title, content, themeVal, username string
		var imagePath sql.NullString
		var userLiked, userDisliked bool
		if err := rows.Scan(&id, &title, &content, &imagePath, &themeVal, &username, &likes, &dislikes, &userLiked, &userDisliked); err != nil {
			return nil, err
		}
		post := Post{
			ID:           id,
			Title:        title,
			Content:      content,
			Theme:        themeVal,
			Username:     username,
			Likes:        likes,
			Dislikes:     dislikes,
			UserLiked:    userLiked,
			UserDisliked: userDisliked,
		}
		if imagePath.Valid {
			post.ImagePath = imagePath.String
		}
		posts = append(posts, post)
	}
	return posts, nil
}
