package backend

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func createPost(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(10 << 20)

	title := r.FormValue("title")
	content := r.FormValue("content")

	file, handler, err := r.FormFile("image")

	if err == nil {

		defer file.Close()

		dst, err := os.Create("uploads/" + handler.Filename)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer dst.Close()

		io.Copy(dst, file)
	}

	fmt.Fprint(w, "Post créé" + title + " " + content)
}