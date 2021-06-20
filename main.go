// webserver for ournotes
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/go-sql-driver/mysql"

	"github.com/muesli/termenv"
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

//wihHe3gxtuBiXMb
type UsersDB struct {
	Login    string
	Password string
	Time     string
	Cookie   string
	Invite   bool
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// validation the user
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
		trueInv := false
		trueTkn := ""

		result, err := myDB.Query("SELECT `cookie`, `invite` FROM auth;")
		checkErr(err)
		defer result.Close()

		u := UsersDB{}

		for result.Next() {
			err := result.Scan(&u.Cookie, &u.Invite)
			checkErr(err)
			if u.Cookie == sessionToken {
				trueTkn = u.Cookie
				trueInv = u.Invite
			}
		}

		if sessionToken != trueTkn {
			//w.WriteHeader(http.StatusUnauthorized)
			http.Redirect(w, r, "/signup", http.StatusSeeOther)
			return
		} else if !trueInv {
			http.Redirect(w, r, "/notfound", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
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

	router.Group(func(r chi.Router) {

		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)

		r.Use(middleware.Timeout(60 * time.Second))

		r.NotFound(notFoundHandler)
		r.Get("/admin", adminHandler)
		r.Post("/invite/{login:[[A-a-Z-z-0-9]+}", inviteHandler)
	})

	server := &http.Server{
		Addr:         *addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	profile := termenv.ColorProfile()
	//fmt.Fprintf(os.Stdout, "%s server is running on %s\n", time.Now().Format("2006/01/02 15:04:05"), os.Getenv("PORT"))
	str := fmt.Sprintf("%s server is running on %s\n", time.Now().Format("2006/01/02 15:04:05"), *addr)
	fmt.Fprint(os.Stdout, termenv.String(str).Bold().Foreground(profile.Color("#71BEF2")))
	//log.Fatal(server.ListenAndServeTLS("yourcert.crt", "yourkey.key"))
	log.Fatal(server.ListenAndServe())
}

func main() {
	db, err := openDB()
	checkErr(err)
	myDB = db
	defer myDB.Close()

	previewRainbow("warning.txt")
	chiStart()
}
