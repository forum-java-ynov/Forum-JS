package backend

import (
	"fmt"
	"net/http"
)

func Server() {
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("frontend/js"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("frontend/css"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/html/index.html")
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {	
		http.ServeFile(w, r, "frontend/html/register.html")
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/html/login.html")
	})

	http.HandleFunc("/db/register", register)
	http.HandleFunc("/db/login", login)
	http.HandleFunc("/db/create_post", createPost)

	fmt.Println("http://localhost:8080")
	http.ListenAndServe(":8080", nil)

}
