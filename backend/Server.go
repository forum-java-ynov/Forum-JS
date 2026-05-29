package backend

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
)

var templates = loadTemplates()

func loadTemplates() *template.Template {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalf("Impossible de récupérer le chemin du fichier source")
	}
	baseDir := filepath.Dir(file)

	paths := []string{
		filepath.Join(baseDir, "..", "frontend", "html", "*.html"),
		filepath.Join(baseDir, "frontend", "html", "*.html"),
		filepath.Join("frontend", "html", "*.html"),
	}

	for _, p := range paths {
		matches, err := filepath.Glob(p)
		if err == nil && len(matches) > 0 {
			tmpl, err := template.ParseGlob(p)
			if err != nil {
				log.Fatalf("Failed to parse templates from %s: %v", p, err)
			}
			return tmpl
		}
	}

	log.Fatalf("Templates not found in paths: %v", paths)
	return nil
}

type IndexData struct {
	Posts []Post
}

func showIndex(w http.ResponseWriter, r *http.Request) {
	posts, err := getPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i := range posts {
		comments, err := getComments(fmt.Sprint(posts[i].ID))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		posts[i].Comments = comments
	}

	data := IndexData{Posts: posts}
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Server() {
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("frontend/js"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("frontend/css"))))
	http.Handle("/frontend/", http.StripPrefix("/frontend/", http.FileServer(http.Dir("frontend"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))
	
	//routes
	http.HandleFunc("/", showIndex)

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/html/register.html")
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/html/login.html")
	})
	//auth google
	http.HandleFunc("/auth/logout", handleLogout)
	http.HandleFunc("/api/me", handleMe)
	http.HandleFunc("/auth/google/login", handleGoogleLogin)
	http.HandleFunc("/auth/google/callback", handleGoogleCallback)

	//appel db
	http.HandleFunc("/db/register", register)
	http.HandleFunc("/db/login", login)
	http.HandleFunc("/db/create_post", createPost)
	http.HandleFunc("/db/posts", showPosts)
	http.HandleFunc("/db/delete_post", deletePostHandler)
	http.HandleFunc("/db/create_commente", createCommente)
	http.HandleFunc("/db/comments", showComments)
	http.HandleFunc("/db/toggle_like", ToggleLikeHandler)
	http.HandleFunc("/db/toggle_comment_like", ToggleCommentLikeHandler)

	fmt.Println("http://localhost:8082")
	http.ListenAndServe(":8082", nil)

}
