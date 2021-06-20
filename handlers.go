package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	result, err := myDB.Query("SELECT * FROM `notes`")
	checkErr(err)
	var cnt int = 1
	allPosts = []ArticleDB{}
	for result.Next() {
		var post ArticleDB
		err := result.Scan(&post.Id, &post.Title, &post.Note, &post.Time)
		checkErr(err)

		post.Count = cnt
		cnt++
		allPosts = append(allPosts, post)
	}

	for i, j := 0, len(allPosts)-1; i < j; i, j = i+1, j-1 {
		allPosts[i], allPosts[j] = allPosts[j], allPosts[i]
	}

	defer result.Close()

	files := []string{
		"html/index.html",
		"html/header.html",
		"html/footer.html",
	}
	tmpl, err := template.ParseFiles(files...)
	checkErr(err)
	tmpl.ExecuteTemplate(w, "index", allPosts)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"html/notfound.html",
	}
	tmpl, err := template.ParseFiles(files...)
	checkErr(err)
	tmpl.Execute(w, nil)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"html/create.html",
		"html/header.html",
		"html/footer.html",
	}
	tmpl, err := template.ParseFiles(files...)
	checkErr(err)
	tmpl.ExecuteTemplate(w, "create", nil)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	note := r.FormValue("note")

	title = strings.TrimSpace(title)
	note = strings.TrimSpace(note)

	if title != "" && note != "" && len([]byte(title)) < 255 && len([]byte(note)) < 65535 {
		data := fmt.Sprintf("INSERT INTO `notes` (`title`, `note`, `time`) VALUES ('%s', '%s', '%s');",
			title, note, time.Now().Format("2006/01/02 15:04:05"))
		_, err := myDB.Exec(data)
		//result, err := db.Query(data)
		checkErr(err)
		//defer result.Close()
	} else {
		http.Redirect(w, r, "/create", http.StatusNoContent)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	val := chi.URLParam(r, "id")

	val = strings.TrimSpace(val)

	if val != "" {
		data := fmt.Sprintf("DELETE FROM `notes` WHERE `id`='%s';", val)
		_, err := myDB.Exec(data)
		//result, err := db.Query(data)
		checkErr(err)
	} else {
		http.Redirect(w, r, "/notfound", http.StatusSeeOther)
	}
	//defer result.Close()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func noteHandler(w http.ResponseWriter, r *http.Request) {
	val := chi.URLParam(r, "id")

	val = strings.TrimSpace(val)

	if val != "" {
		w.WriteHeader(http.StatusOK)

		result, err := myDB.Query(fmt.Sprintf("SELECT * FROM `notes` WHERE `id` = %s;", val))
		checkErr(err)

		note := ArticleDB{}

		for result.Next() {
			err := result.Scan(&note.Id, &note.Title, &note.Note, &note.Time)
			checkErr(err)
		}

		defer result.Close()

		files := []string{
			"html/note.html",
			"html/header.html",
			"html/footer.html",
		}

		tmpl, err := template.ParseFiles(files...)
		checkErr(err)
		tmpl.ExecuteTemplate(w, "note", note)
	} else {
		http.Redirect(w, r, "/notfound", http.StatusFound)
		return
	}
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	result, err := myDB.Query("SELECT * FROM `notes`")
	checkErr(err)
	var cnt int = 1
	allPosts = []ArticleDB{}
	for result.Next() {
		var post ArticleDB
		err := result.Scan(&post.Id, &post.Title, &post.Note, &post.Time)
		checkErr(err)
		post.Count = cnt
		cnt++
		allPosts = append(allPosts, post)
	}
	defer result.Close()
	b, err := json.Marshal(allPosts)
	checkErr(err)
	w.Write(b)
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"html/signin.html",
		"html/header.html",
		"html/footer.html",
	}
	tmpl, err := template.ParseFiles(files...)
	checkErr(err)
	tmpl.ExecuteTemplate(w, "signin", nil)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"html/signup.html",
		"html/header.html",
		"html/footer.html",
	}
	tmpl, err := template.ParseFiles(files...)
	checkErr(err)
	tmpl.ExecuteTemplate(w, "signup", nil)
}

func regHandler(w http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	pass := r.FormValue("pass")

	login = strings.TrimSpace(login)
	pass = strings.TrimSpace(pass)

	if login != "" && pass != "" && len([]byte(login)) < 100 && len([]byte(pass)) < 255 {
		//db, err := sql.Open("mysql", signDB)
		//checkErr(err)
		//defer db.Close()
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), 8)
		checkErr(err)
		data := fmt.Sprintf("INSERT INTO `auth` (`login`, `pass`, `time`, `cookie`, `invite`) VALUES ('%s', '%s', '%s', '%s', %v);",
			login, hashedPassword, time.Now().Format("2006/01/02 15:04:05"), "0", false)
		result, err := myDB.Query(data)
		checkErr(err)
		defer result.Close()
	} else {
		http.Redirect(w, r, "/signup", http.StatusNoContent)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("login")
	pass := r.FormValue("pass")

	username = strings.TrimSpace(username)
	pass = strings.TrimSpace(pass)
	//fmt.Println(username, pass)
	if username != "" && pass != "" && len([]byte(username)) < 100 && len([]byte(pass)) < 255 {
		valUser := validUserFromDB(username)

		// Get the expected password from our in memory map
		expectedPassword, ok := valUser[username]

		// If a password exists for the given user
		// AND, if it is the same as the password we received, the we can move ahead
		// if NOT, then we return an "Unauthorized" status
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else if err := bcrypt.CompareHashAndPassword([]byte(expectedPassword), []byte(pass)); err != nil {
			// Compare the stored hashed password, with the hashed version of the password that was received
			// If the two passwords don't match, return a 401 status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Create a new random session token
		sessionToken := uuid.NewV4().String()

		data := fmt.Sprintf("UPDATE `auth` SET `cookie`='%s' WHERE `login`='%s';", sessionToken, username)
		result, err := myDB.Query(data)
		checkErr(err)
		defer result.Close()

		// Finally, we set the client cookie for "session_token" as the session token we just generated
		// we also set an expiry time of 12 hours, the same as the cache
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: time.Now().Add(12 * time.Hour),
			Path:    "/",
		})
	} else {
		http.Redirect(w, r, "/signup", http.StatusNoContent)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// admin panel for invite users (only for login 'admin')
func adminHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			//w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/signup", http.StatusSeeOther)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	result, err := myDB.Query("SELECT `cookie` FROM auth WHERE `login`='admin';")
	checkErr(err)
	defer result.Close()

	u := UsersDB{}

	for result.Next() {
		err := result.Scan(&u.Cookie)
		checkErr(err)
		if u.Cookie != sessionToken {
			http.Redirect(w, r, "/signup", http.StatusSeeOther)
			return
		}
	}

	listNotInvited, err := myDB.Query("SELECT `login`, `time`, `invite` FROM auth WHERE `invite`=false;")
	checkErr(err)
	defer listNotInvited.Close()
	notInvited := []UsersDB{}

	for listNotInvited.Next() {
		var user UsersDB
		err := listNotInvited.Scan(&user.Login, &user.Time, &user.Invite)
		checkErr(err)
		if !user.Invite {
			notInvited = append(notInvited, user)
		}
	}

	files := []string{
		"html/admin.html",
		"html/header.html",
		"html/footer.html",
	}

	tmpl, err := template.ParseFiles(files...)
	checkErr(err)
	tmpl.ExecuteTemplate(w, "admin", notInvited)
}

// func invite a user
func inviteHandler(w http.ResponseWriter, r *http.Request) {
	val := chi.URLParam(r, "login")

	val = strings.TrimSpace(val)

	if val != "" {
		data := fmt.Sprintf("UPDATE `auth` SET `invite`=%v WHERE `login`='%s'", true, val)
		_, err := myDB.Exec(data)
		//result, err := db.Query(data)
		checkErr(err)
	} else {
		http.Redirect(w, r, "/notfound", http.StatusSeeOther)
	}
	//defer result.Close()
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
