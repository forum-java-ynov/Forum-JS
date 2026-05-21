package backend

import (
	"fmt"
	"net/http"
)

func Server() {
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("frontend/js"))))

	http.Handle("/frontend/", http.StripPrefix("/frontend/", http.FileServer(http.Dir("frontend"))))

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

	fmt.Println("http://localhost:8082")
	http.ListenAndServe(":8082", nil)

}
