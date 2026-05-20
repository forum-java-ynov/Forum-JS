package backend

import (
	"fmt"
	"net/http"
)

func Server() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/html/index.html")
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {	
		http.ServeFile(w, r, "../frontend/html/register.html")
	})
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../frontend/html/login.html")
	})

	fmt.Println("http://localhost:8080")
	http.ListenAndServe(":8080", nil)

}