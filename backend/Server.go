package backend

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

var templates = loadTemplates()

type ErrorData struct {
	Code    int
	Message string
}

var errorMessages = map[int]string{
	http.StatusBadRequest:          "Invalid Request",
	http.StatusUnauthorized:        "You need to be logged in",
	http.StatusForbidden:           "Permission needed",
	http.StatusNotFound:            "Page not Found",
	http.StatusMethodNotAllowed:    "Unauthorized method",
	http.StatusInternalServerError: "Internal error Occurred",
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
		log.Fatalf("Can't resolve source code path")
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
		httpError(w, http.StatusNotFound)
		return
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	currentUserID, _ := getCurrentUserID(w, r)

	var posts []Post
	var err error

	themeFilter := r.URL.Query().Get("theme")
	likeFilter := r.URL.Query().Get("post_liked")
	myPostsFilter := r.URL.Query().Get("hispost")

	var userID int
	if likeFilter != "" || myPostsFilter != "" {
		userID, err = getCurrentUserID(w, r)
		if err != nil || userID == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
	}

	posts, err = getFilteredPosts(themeFilter, currentUserID, likeFilter != "", myPostsFilter != "")

	if err != nil {
		log.Println("Erreur lors de la récupération des posts:", err)
		serverError(w)
		return
	}

	for i := range posts {
		comments, err := getComments(posts[i].ID, currentUserID)
		if err != nil {
			log.Println("Erreur lors de la récupération des commentaires:", err)
			serverError(w)
			return
		}
		posts[i].Comments = comments
	}

	data := IndexData{Posts: posts}

	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Println("Erreur template:", err)
		return
	}
}

func Server() {
	go cleanupVisitors()
	InitDB()

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
		httpError(w, http.StatusNotFound)
	})

	//auth github
	http.HandleFunc("/auth/github/login", handleGitHubLogin)
	http.HandleFunc("/auth/github/callback", handleGitHubCallback)

	// auth google
	http.HandleFunc("/auth/logout", handleLogout)
	http.HandleFunc("/api/me", isAuthenticated(handleMe))
	http.HandleFunc("/auth/google/login", handleGoogleLogin)
	http.HandleFunc("/auth/google/callback", handleGoogleCallback)

	// error tests
	http.HandleFunc("/test500", func(w http.ResponseWriter, r *http.Request) {
		serverError(w)
	})
	http.HandleFunc("/test403", func(w http.ResponseWriter, r *http.Request) {
		httpError(w, http.StatusForbidden)
	})
	http.HandleFunc("/test401", func(w http.ResponseWriter, r *http.Request) {
		httpError(w, http.StatusUnauthorized)
	})

	// Routes DataBase
	http.HandleFunc("/db/register", rateLimiter(register, 5, time.Minute))
	http.HandleFunc("/db/login", rateLimiter(login, 5, time.Minute))
	http.HandleFunc("/db/create_post", isAuthenticated(createPostHandler))
	http.HandleFunc("/db/posts", showPostsHandler)
	http.HandleFunc("/db/delete_post", isAuthenticated(deletePostHandler))
	http.HandleFunc("/db/edit-post", isAuthenticated(editPostHandler))

	http.HandleFunc("/db/create_comment", isAuthenticated(createCommentHandler))
	http.HandleFunc("/db/comments", showCommentsHandler)
	http.HandleFunc("/db/edit_comment", isAuthenticated(editCommentHandler))
	http.HandleFunc("/db/delete_comment", isAuthenticated(deleteCommentHandler))

	http.HandleFunc("/db/toggle_like", isAuthenticated(ToggleLikeHandler))
	http.HandleFunc("/db/toggle_dislike", isAuthenticated(ToggleDislikeHandler))
	http.HandleFunc("/db/toggle_comment_like", isAuthenticated(ToggleCommentLikeHandler))
	http.HandleFunc("/db/toggle_comment_dislike", isAuthenticated(ToggleCommentDislikeHandler))

	fmt.Println("http://localhost:8082")
	http.ListenAndServe(":8082", nil)

	//shutdown server
	srv := &http.Server{Addr: ":8082"}

	go func() {
		fmt.Println("Server running at http://localhost:8082")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Erreur du serveur: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("\nShutting down server...")

	// closing db
	if DB != nil {
		DB.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("stopped forced %v", err)
	}

	fmt.Println("server shut down successfully")
}
