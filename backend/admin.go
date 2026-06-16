package backend

import (
	"log"
	"net/http"
	"strconv"
)

type AdminUser struct {
	ID       int
	Username string
	Email    string
	Role     string
}

type AdminPost struct {
	ID       int
	Title    string
	Username string
	Theme    string
	Likes    int
	Dislikes int
	Comments int
}

type AdminStats struct {
	TotalUsers    int
	TotalPosts    int
	TotalComments int
	TotalLikes    int
}

type AdminData struct {
	Users []AdminUser
	Posts []AdminPost
	Stats AdminStats
}

func isAdmin(w http.ResponseWriter, r *http.Request) bool {
	userID, err := getCurrentUserID(w, r)
	if err != nil || userID == 0 {
		return false
	}
	admin, err := getUserRole(userID)
	if err != nil {
		return false
	}
	return admin
}

func showAdminPanel(w http.ResponseWriter, r *http.Request) {
	if !isAdmin(w, r) {
		httpError(w, http.StatusForbidden)
		return
	}

	users, err := getAllUsers()
	if err != nil {
		log.Println("admin: getAllUsers:", err)
		serverError(w)
		return
	}

	posts, err := getAdminPosts()
	if err != nil {
		log.Println("admin: getAdminPosts:", err)
		serverError(w)
		return
	}

	stats, err := getAdminStats()
	if err != nil {
		log.Println("admin: getAdminStats:", err)
		serverError(w)
		return
	}

	data := AdminData{Users: users, Posts: posts, Stats: stats}
	if err := templates.ExecuteTemplate(w, "admin.html", data); err != nil {
		log.Println("admin: template error:", err)
	}
}

func adminDeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed)
		return
	}
	if !isAdmin(w, r) {
		httpError(w, http.StatusForbidden)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	if err := deleteUserByID(id); err != nil {
		log.Println("admin: deleteUserByID:", err)
		serverError(w)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func adminDeletePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed)
		return
	}
	if !isAdmin(w, r) {
		httpError(w, http.StatusForbidden)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	DB.Exec("DELETE FROM comment_like WHERE comments_id IN (SELECT id FROM comments WHERE post_id = ?);", id)
	DB.Exec("DELETE FROM comment_dislike WHERE comments_id IN (SELECT id FROM comments WHERE post_id = ?);", id)
	DB.Exec("DELETE FROM comments WHERE post_id = ?;", id)
	DB.Exec("DELETE FROM post_like WHERE post_id = ?;", id)
	DB.Exec("DELETE FROM post_dislike WHERE post_id = ?;", id)
	DB.Exec("DELETE FROM posts WHERE id = ?;", id)

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func adminPromoteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed)
		return
	}
	if !isAdmin(w, r) {
		httpError(w, http.StatusForbidden)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httpError(w, http.StatusBadRequest)
		return
	}

	if err := toggleUserRole(id); err != nil {
		log.Println("admin: toggleUserRole:", err)
		serverError(w)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
