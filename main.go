package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("your-secret-key"))

func isAuthenticated(r *http.Request) bool {
	session, err := store.Get(r, "session-name")
	if err != nil {
		return false
	}
	return session.Values["user_email"] != nil
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	title := r.URL.Path[len("/delete/"):]
	err := os.Remove(filepath.Join("data", title+".txt"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	http.HandleFunc("/delete/", deleteHandler)
}
