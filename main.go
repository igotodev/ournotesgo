// server for [ournotes] site
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/go-sql-driver/mysql"
)

const signDB string = "root:password@tcp(localhost:3306)/yourdb" // example

type ArticleDB struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Note  string `json:"note"`
	Time  string `json:"time"`
}

var allPosts = []ArticleDB{}

//for basic auth
var admins = map[string]string{
	"admin": "password", // example
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", signDB)
	checkErr(err)
	defer db.Close()
	result, err := db.Query("SELECT * FROM `notes`")
	checkErr(err)

	allPosts = []ArticleDB{}
	for result.Next() {
		var post ArticleDB
		err := result.Scan(&post.Id, &post.Title, &post.Note, &post.Time)
		checkErr(err)

		allPosts = append(allPosts, post)
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
	//fmt.Fprintln(w, title+" "+note)
	if strings.TrimSpace(title) != "" && strings.TrimSpace(note) != "" {
		db, err := sql.Open("mysql", signDB)
		checkErr(err)
		defer db.Close()
		data := fmt.Sprintf("INSERT INTO `notes` (`title`, `note`, `time`) VALUES ('%s', '%s', '%s');", title, note, time.Now().Format("2006/01/02 15:04:05"))
		result, err := db.Query(data)
		checkErr(err)
		defer result.Close()
	} else {
		http.Redirect(w, r, "/create", http.StatusNoContent)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func noteHandler(w http.ResponseWriter, r *http.Request) {
	val := chi.URLParam(r, "id")
	if val != "" {
		w.WriteHeader(http.StatusOK)

		db, err := sql.Open("mysql", signDB)
		checkErr(err)
		defer db.Close()
		result, err := db.Query(fmt.Sprintf("SELECT * FROM `notes` WHERE `id` = %s;", val))

		note := ArticleDB{}

		for result.Next() {
			err := result.Scan(&note.Id, &note.Title, &note.Note, &note.Time)
			//fmt.Println(note)
			checkErr(err)
		}

		checkErr(err)
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

/*
// middleware for user Auth
func myMiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//code will be here
		next.ServeHTTP(w, r)
	})
}
*/

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", signDB)
	checkErr(err)
	defer db.Close()
	result, err := db.Query("SELECT * FROM `notes`")
	checkErr(err)
	allPosts = []ArticleDB{}
	for result.Next() {
		var post ArticleDB
		err := result.Scan(&post.Id, &post.Title, &post.Note, &post.Time)
		checkErr(err)

		allPosts = append(allPosts, post)
	}
	defer result.Close()
	b, err := json.Marshal(allPosts)
	checkErr(err)
	w.Write(b)
}

func chiStart() {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	//router.Use(myMiddlewareAuth)
	router.Use(middleware.BasicAuth("Enter your login and password: ", admins))
	router.Use(middleware.Timeout(60 * time.Second))

	router.NotFound(notFoundHandler)

	router.Get("/", indexHandler)
	router.Get("/create", createHandler)
	router.Get("/note/{id:[0-9]+}", noteHandler)
	router.Get("/json", jsonHandler)

	router.Post("/save-art", saveHandler)

	router.Mount("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	server := &http.Server{
		//Addr:         ":" + os.Getenv("PORT"),
		Addr:         ":8181",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	//fmt.Fprintf(os.Stdout, "%s server is running on port %s\n", time.Now().Format("2006/01/02 15:04:05"), os.Getenv("PORT"))
	fmt.Fprintf(os.Stdout, "%s server is running on port 8181\n", time.Now().Format("2006/01/02 15:04:05"))
	log.Fatal(server.ListenAndServe())
}

func main() {
	chiStart()
}
