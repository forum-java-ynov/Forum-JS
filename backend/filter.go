package backend

import (
	"encoding/json"
	"net/http"
	"database/sql"
)

func filterPostsHandler(w http.ResponseWriter, r *http.Request) {
	theme := r.URL.Query().Get("theme")
	posts, err := filterPostsByTheme(theme)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func filterPostsByTheme(theme string) ([]Post, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	query := `
		SELECT posts.id, posts.title, posts.content, posts.theme,
		       COALESCE(users.full_name, users.username, 'Utilisateur')
		FROM posts
		LEFT JOIN users ON posts.user_id = users.id
	`
	var args []interface{}
	if theme != "" {
		query += " WHERE posts.theme = ?"
		args = append(args, theme)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Theme, &post.Username); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}