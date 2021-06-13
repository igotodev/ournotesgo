// server for [ournotes] site
package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
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

var myDB *sql.DB

type ArticleDB struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Note  string `json:"note"`
	Time  string `json:"time"`
}

var allPosts = []ArticleDB{}

type UsersDB struct {
	Login    string
	Password string
	Time     string
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	result, err := myDB.Query("SELECT * FROM `notes`")
	checkErr(err)

	allPosts = []ArticleDB{}
	for result.Next() {
		var post ArticleDB
		err := result.Scan(&post.Id, &post.Title, &post.Note, &post.Time)
		checkErr(err)

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

	if strings.TrimSpace(title) != "" && strings.TrimSpace(note) != "" && len([]byte(title)) < 255 && len([]byte(note)) < 255 {
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
	result, err := myDB.Query("SELECT * FROM `notes`")
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

func validUsersFromDB() map[string]string {
	result, err := myDB.Query("SELECT `login`, `pass` FROM `auth`;")
	checkErr(err)
	lp := make(map[string]string)
	var users []UsersDB
	for result.Next() {
		var u UsersDB
		err := result.Scan(&u.Login, &u.Password)
		checkErr(err)
		users = append(users, u)
	}
	defer result.Close()
	for _, v := range users {
		lp[v.Login] = v.Password
	}
	return lp
}

// consolePrint printed UTF8 text from file to os.Stdout (not necessarily, it's for fun)
func consolePrint(file string) {
	logo, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(150 * time.Millisecond)
	scanner := bufio.NewScanner(logo)
	for scanner.Scan() {
		myBytes := scanner.Text() + "\n"
		for _, v := range myBytes {
			time.Sleep(25 * time.Millisecond)
			fmt.Fprint(os.Stdout, string(v))
		}
	}
	time.Sleep(150 * time.Millisecond)
	fmt.Fprintf(os.Stdout, "\n")
}

func openDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", signDB)
	if err != nil {
		return nil, err
	}
	return db, err
}

func chiStart() {
	addr := flag.String("addr", ":80", "host address")
	flag.Parse()

	consolePrint("warning.txt")

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	//router.Use(myMiddlewareAuth)
	router.Use(middleware.BasicAuth("Enter your login and password: ", validUsersFromDB()))
	router.Use(middleware.Timeout(60 * time.Second))

	router.NotFound(notFoundHandler)

	router.Get("/", indexHandler)
	router.Get("/create", createHandler)
	router.Get("/note/{id:[0-9]+}", noteHandler)
	router.Get("/json", jsonHandler)

	router.Post("/save-art", saveHandler)
	router.Post("/delete/{id:[0-9]+}", deleteHandler)

	router.Mount("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	server := &http.Server{
		Addr:         *addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	//fmt.Fprintf(os.Stdout, "%s server is running on %s\n", time.Now().Format("2006/01/02 15:04:05"), os.Getenv("PORT"))
	fmt.Fprintf(os.Stdout, "%s server is running on %s\n", time.Now().Format("2006/01/02 15:04:05"), *addr)
	//log.Fatal(server.ListenAndServeTLS("yourcert.crt", "yourkey.key"))
	log.Fatal(server.ListenAndServe())
}

func main() {
	db, err := openDB()
	checkErr(err)
	myDB = db
	defer myDB.Close()

	chiStart()
}
