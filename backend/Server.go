package backend

import (
	"fmt"
	"net/http"
)

func Server() {
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("frontend/js"))))
	http.Handle("/frontend/", http.StripPrefix("/frontend/", http.FileServer(http.Dir("frontend"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))

	//routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/html/index.html")
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/html/register.html")
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/html/login.html")
	})
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

	fmt.Println("http://localhost:8082")
	http.ListenAndServe(":8082", nil)

}
