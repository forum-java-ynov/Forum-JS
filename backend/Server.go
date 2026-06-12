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

type ErrorData struct {
	Code    int
	Message string
}

var errorMessages = map[int]string{
	http.StatusBadRequest:          "Requête invalide",
	http.StatusUnauthorized:        "Vous devez être connecté",
	http.StatusForbidden:           "Vous n'avez pas la permission",
	http.StatusNotFound:            "Page introuvable",
	http.StatusMethodNotAllowed:    "Méthode non autorisée",
	http.StatusInternalServerError: "Une erreur interne est survenue",
}

func httpError(w http.ResponseWriter, code int) {
	msg, ok := errorMessages[code]
	if !ok {
		msg = "Une erreur est survenue"
	}
	w.WriteHeader(code)
	if err := templates.ExecuteTemplate(w, "error.html", ErrorData{Code: code, Message: msg}); err != nil {
		log.Println("httpError template error:", err)
		fmt.Fprintf(w, "%d - %s", code, msg)
	}
}

func serverError(w http.ResponseWriter) {
	httpError(w, http.StatusInternalServerError)
}

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
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		http.ServeFile(w, r, "frontend/html/404.html")
		return
	}
	posts, err := getPosts()

	themeFilter := r.URL.Query().Get("theme")
	if themeFilter != "" {
		posts, err = filterPostsByTheme(themeFilter)
	} else {
		posts, err = getPosts()
	}

	if err != nil {
		log.Println(err)
		serverError(w)
		return
	}

	for i := range posts {
		comments, err := getComments(fmt.Sprint(posts[i].ID))
		if err != nil {
			log.Println(err)
			serverError(w)
			return
		}
		posts[i].Comments = comments
	}

	data := IndexData{Posts: posts}
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Println(err)
		serverError(w)
	}
}

func Server() {
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("frontend/js"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("frontend/css"))))
	http.Handle("/frontend/", http.StripPrefix("/frontend/", http.FileServer(http.Dir("frontend"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))

	http.HandleFunc("/", showIndex)
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/html/register.html")
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/html/login.html")
	})
	http.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/html/404.html")
	})
	//auth google
	http.HandleFunc("/auth/logout", handleLogout)
	http.HandleFunc("/api/me", isAuthenticated(handleMe))
	http.HandleFunc("/auth/google/login", handleGoogleLogin)
	http.HandleFunc("/auth/google/callback", handleGoogleCallback)

	http.HandleFunc("/test500", func(w http.ResponseWriter, r *http.Request) {
		serverError(w)
	})
	http.HandleFunc("/test403", func(w http.ResponseWriter, r *http.Request) {
		httpError(w, http.StatusForbidden)
	})
	http.HandleFunc("/test401", func(w http.ResponseWriter, r *http.Request) {
		httpError(w, http.StatusUnauthorized)
	})

	http.HandleFunc("/db/register", register)
	http.HandleFunc("/db/login", login)
	http.HandleFunc("/db/create_post", isAuthenticated(createPost))
	http.HandleFunc("/db/posts", showPosts)
	http.HandleFunc("/db/delete_post", isAuthenticated(deletePostHandler))
	http.HandleFunc("/db/delete_comment", isAuthenticated(deleteCommentHandler))
	http.HandleFunc("/db/create_commente", isAuthenticated(createCommente))
	http.HandleFunc("/db/comments", showComments)
	http.HandleFunc("/db/edit_commente", isAuthenticated(editComment))
	http.HandleFunc("/db/toggle_like", isAuthenticated(ToggleLikeHandler))
	http.HandleFunc("/db/toggle_dislike", isAuthenticated(ToggleDislikeHandler))
	http.HandleFunc("/db/toggle_comment_like", isAuthenticated(ToggleCommentLikeHandler))
	http.HandleFunc("/db/toggle_comment_dislike", isAuthenticated(ToggleCommentDislikeHandler))

	fmt.Println("http://localhost:8082")
	http.ListenAndServe(":8082", nil)
}
