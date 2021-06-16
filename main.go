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
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
)

const signDB string = "root:password@tcp(localhost:3306)/yourdb" // example

var myDB *sql.DB

type ArticleDB struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Note  string `json:"note"`
	Time  string `json:"time"`
	Count int    `json:"count"`
}

var allPosts = []ArticleDB{}

type UsersDB struct {
	Login     string
	Password  string
	Time      string
	Cookie    string
	NewCookie string
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

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

// middleware for users cookies
func checkCookiesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		trueTkn := ""

		result, err := myDB.Query("SELECT `cookie` FROM `auth`;")
		checkErr(err)
		defer result.Close()

		u := UsersDB{}

		for result.Next() {
			err := result.Scan(&u.Cookie)
			checkErr(err)
			if u.Cookie == sessionToken {
				trueTkn = u.Cookie
			}
		}

		if sessionToken != trueTkn {
			//w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/signup", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
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

//----------------------------
func validUserFromDB(login string) map[string]string {
	data := fmt.Sprintf("SELECT `pass` FROM `auth` WHERE `login`='%s';", login)
	result, err := myDB.Query(data)
	checkErr(err)
	lp := make(map[string]string)
	var user UsersDB
	for result.Next() {
		err := result.Scan(&user.Password)
		checkErr(err)
		lp[login] = user.Password
	}
	defer result.Close()

	return lp
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

	if strings.TrimSpace(login) != "" && strings.TrimSpace(pass) != "" {
		//db, err := sql.Open("mysql", signDB)
		//checkErr(err)
		//defer db.Close()
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), 8)
		checkErr(err)
		data := fmt.Sprintf("INSERT INTO `auth` (`login`, `pass`, `time`, `cookie`, `newcookie`) VALUES ('%s', '%s', '%s', '%s', '%s');",
			login, hashedPassword, time.Now().Format("2006/01/02 15:04:05"), "0", "0")
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
	//fmt.Println(username, pass)
	if strings.TrimSpace(username) != "" && strings.TrimSpace(pass) != "" && len([]byte(username)) < 255 && len([]byte(pass)) < 255 {
		valUser := validUserFromDB(strings.TrimSpace(username))

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

// consolePrint printed UTF8 text from file to os.Stdout (not necessarily, it's for fun)
func consolePrint(file string) {
	logo, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)
	scanner := bufio.NewScanner(logo)
	for scanner.Scan() {
		myBytes := scanner.Text() + "\n"
		for _, v := range myBytes {
			time.Sleep(25 * time.Millisecond)
			fmt.Fprint(os.Stdout, string(v))
		}
	}
	time.Sleep(100 * time.Millisecond)
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

	router.Group(func(r chi.Router) {

		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)

		r.Use(checkCookiesMiddleware)

		r.Use(middleware.Timeout(60 * time.Second))

		r.NotFound(notFoundHandler)
		r.Get("/", indexHandler)
		r.Get("/create", createHandler)
		r.Get("/note/{id:[0-9]+}", noteHandler)
		r.Get("/json", jsonHandler)

		r.Post("/save-art", saveHandler)
		r.Post("/delete/{id:[0-9]+}", deleteHandler)

	})

	router.Group(func(r chi.Router) {
		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)

		r.Use(middleware.Timeout(60 * time.Second))

		r.Get("/signup", signupHandler)
		r.Get("/signin", signinHandler)
		r.Post("/reg", regHandler)
		r.Post("/auth", authHandler)

		r.Mount("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	})

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
