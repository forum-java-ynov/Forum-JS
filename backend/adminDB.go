package backend

func getAllUsers() ([]AdminUser, error) {
	rows, err := DB.Query(`
		SELECT id, username, email, role FROM users ORDER BY id ASC;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []AdminUser
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func getAdminPosts() ([]AdminPost, error) {
	rows, err := DB.Query(`
		SELECT
			posts.id,
			posts.title,
			COALESCE(users.username, 'Inconnu') as username,
			COALESCE(posts.theme, ''),
			COUNT(DISTINCT post_like.id) as likes,
			COUNT(DISTINCT post_dislike.id) as dislikes,
			COUNT(DISTINCT comments.id) as comments
		FROM posts
		LEFT JOIN users ON posts.user_id = users.id
		LEFT JOIN post_like ON posts.id = post_like.post_id
		LEFT JOIN post_dislike ON posts.id = post_dislike.post_id
		LEFT JOIN comments ON posts.id = comments.post_id
		GROUP BY posts.id
		ORDER BY posts.id DESC;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []AdminPost
	for rows.Next() {
		var p AdminPost
		if err := rows.Scan(&p.ID, &p.Title, &p.Username, &p.Theme, &p.Likes, &p.Dislikes, &p.Comments); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func getAdminStats() (AdminStats, error) {
	var s AdminStats
	DB.QueryRow("SELECT COUNT(*) FROM users;").Scan(&s.TotalUsers)
	DB.QueryRow("SELECT COUNT(*) FROM posts;").Scan(&s.TotalPosts)
	DB.QueryRow("SELECT COUNT(*) FROM comments;").Scan(&s.TotalComments)
	DB.QueryRow("SELECT COUNT(*) FROM post_like;").Scan(&s.TotalLikes)
	return s, nil
}

func deleteUserByID(id int) error {
	// clean up everything owned by the user first
	DB.Exec("DELETE FROM comment_like WHERE comments_id IN (SELECT id FROM comments WHERE user_id = ?);", id)
	DB.Exec("DELETE FROM comment_dislike WHERE comments_id IN (SELECT id FROM comments WHERE user_id = ?);", id)
	DB.Exec("DELETE FROM comments WHERE user_id = ?;", id)
	DB.Exec("DELETE FROM post_like WHERE user_id = ?;", id)
	DB.Exec("DELETE FROM post_dislike WHERE user_id = ?;", id)

	// delete posts owned by user (with their comments)
	rows, _ := DB.Query("SELECT id FROM posts WHERE user_id = ?;", id)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var postID int
			rows.Scan(&postID)
			DB.Exec("DELETE FROM comment_like WHERE comments_id IN (SELECT id FROM comments WHERE post_id = ?);", postID)
			DB.Exec("DELETE FROM comment_dislike WHERE comments_id IN (SELECT id FROM comments WHERE post_id = ?);", postID)
			DB.Exec("DELETE FROM comments WHERE post_id = ?;", postID)
			DB.Exec("DELETE FROM post_like WHERE post_id = ?;", postID)
			DB.Exec("DELETE FROM post_dislike WHERE post_id = ?;", postID)
		}
	}

	DB.Exec("DELETE FROM posts WHERE user_id = ?;", id)
	DB.Exec("DELETE FROM user_sessions WHERE user_id = ?;", id)
	_, err := DB.Exec("DELETE FROM users WHERE id = ?;", id)
	return err
}

func toggleUserRole(id int) error {
	_, err := DB.Exec(`
		UPDATE users SET role = CASE WHEN role = 'admin' THEN 'user' ELSE 'admin' END
		WHERE id = ?;
	`, id)
	return err
}
