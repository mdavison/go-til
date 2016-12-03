package main

import (
	"net/http"
	"html/template"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"

	"encoding/json"

	"github.com/urfave/negroni"
	"github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	gmux "github.com/gorilla/mux"
	"strconv"
	"golang.org/x/crypto/bcrypt"
	"fmt"
	"time"
)

type Page struct {
	Tils []Til
	User string
}

type LoginPage struct {
	Register bool
	Error string
}

type Til struct {
	ID int
	Title string
	Date string
}

type User struct {
	ID int
	Email string
	Password   []byte
}

var db *sql.DB

func verifyDatabase(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if err := db.Ping(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	next(w, r)
}

func NewUser(email, password string) *User {
	pw, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return &User {
		ID: -1,
		Email: email,
		Password: pw,
	}
}

func getStringFromSession(r *http.Request, key string) string {
	var strVal string
	if val := sessions.GetSession(r).Get(key); val != nil {
		strVal = val.(string)
	}

	return strVal
}

func verifyUser(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.URL.Path == "/login" {
		next(w, r)
		return
	}
	if email := getStringFromSession(r, "User"); email != "" {
		var user User
		err := db.QueryRow("SELECT id, email, password FROM users WHERE email = ?", r.FormValue("email")).Scan(&user.ID, &user.Email, &user.Password)
		if err == nil {
			next(w, r)
			return
		}
	}
	//http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	next(w, r)
}

func (til *Til) formatDate() {
	format := "2006-01-02 15:04:05-07:00"

	if tilTime, err := time.Parse(format, til.Date); err == nil {
		til.Date = tilTime.Format("Jan 2 2006")
	}
}

func formatDate(date time.Time) string {
	// 2016-12-03 15:22:41.822142097 -0800 PST
	format := "2006-01-02 15:04:05 -0700 MST"

	if t, err := time.Parse(format, time.Time.String(date)); err ==  nil {
		return t.Format("Jan 2 2006")
	}

	return time.Time.String(date)
}

func main() {
	templates := template.Must(template.ParseFiles(
		"templates/index.html",
		"templates/login.html"))

	db, _ = sql.Open("sqlite3", "til.db")

	mux := gmux.NewRouter()

	s := http.StripPrefix("/resources/", http.FileServer(http.Dir("./resources/")))
	mux.PathPrefix("/resources/").Handler(s)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		email := getStringFromSession(r, "User")
		p := Page{
			Tils: []Til{},
			User: email,
		}
		// Fetch user
		var user User
		err := db.QueryRow("SELECT id, email FROM users WHERE email = ?", email).Scan(&user.ID, &user.Email)

		// Fetch TILs for user
		if err == nil {
			rows, _ := db.Query("SELECT id, title, date FROM tils WHERE user_id = ?", user.ID)
			for rows.Next() {
				var t Til
				rows.Scan(&t.ID, &t.Title, &t.Date)
				t.formatDate()
				p.Tils = append(p.Tils, t)
			}
		}

		if err := templates.ExecuteTemplate(w, "index.html", p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request){
		var page LoginPage
		page.Register = true

		// If user already in session, redirect to home page
		if email := getStringFromSession(r, "User"); email != "" {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		}

		// Using the login.html template for both register and login
		if err := templates.ExecuteTemplate(w, "login.html", page); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request){
		var page LoginPage
		var user User
		validationPasses := true

		// If user already in session, redirect to home page
		if email := getStringFromSession(r, "User"); email != "" {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		}

		if r.FormValue("register") != "" {
			page.Register = true
			// Validation
			if r.FormValue("password") != r.FormValue("password-confirm") {
				page.Error = "Passwords don't match."
				validationPasses = false
			}
			user := NewUser(r.FormValue("email"), r.FormValue("password"))

			// Check if user already exists
			err := db.QueryRow("SELECT id, email, password FROM users WHERE email = ?", r.FormValue("email")).Scan(&user.ID, &user.Email, &user.Password)
			// If no error, user exists: don't add
			if err == nil {
				page.Error = "User already exists"
				validationPasses = false
			}

			if validationPasses {
				if _, err := db.Exec("INSERT INTO users (id, email, password) values (?, ?, ?)", nil, user.Email, user.Password); err != nil {
					page.Error = err.Error()
				} else {
					// Put user into session
					sessions.GetSession(r).Set("User", user.Email)
					http.Redirect(w, r, "/", http.StatusFound)
					return
				}
			}

		} else if r.FormValue("login") != "" {
			fmt.Println("Logging in")
			err := db.QueryRow("SELECT id, email, password FROM users WHERE email = ?", r.FormValue("email")).Scan(&user.ID, &user.Email, &user.Password)
			if err != nil {
				page.Error = "Email/password combination incorrect"
			} else {
				fmt.Println(string(user.Password))

				if err = bcrypt.CompareHashAndPassword(user.Password, []byte(r.FormValue("password"))); err != nil {
					page.Error = "Email/password combination incorrect"
				} else {
					// Put user into session
					sessions.GetSession(r).Set("User", user.Email)
					http.Redirect(w, r, "/", http.StatusFound)
					return
				}
			}
		}

		if err := templates.ExecuteTemplate(w, "login.html", page); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request){
		sessions.GetSession(r).Set("User", nil)

		http.Redirect(w, r, "/login", http.StatusFound)
	})

	mux.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request){
		// Get user
		var user User
		email := getStringFromSession(r, "User")
		err := db.QueryRow("SELECT id, email, password FROM users WHERE email = ?", email).Scan(&user.ID, &user.Email, &user.Password)

		now := time.Now()
		formattedDate := formatDate(now)
		encoder := json.NewEncoder(w)

		// Only insert TIL if we have a user
		if err == nil {
			title := r.FormValue("title")


			row, err := db.Exec("INSERT INTO tils (id, title, user_id, date) values (?, ?, ?, ?)", nil, title, user.ID, now)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			id, _ := row.LastInsertId()
			results := []Til{
				Til{ID: int(id), Title: title, Date: formattedDate},
			}

			if err = encoder.Encode(results); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			results := []Til{
				Til{ID: int(-1), Title: r.FormValue("title"), Date: formattedDate},
			}
			encoder.Encode(results)
		}
	})

	mux.HandleFunc("/delete/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(gmux.Vars(r)["id"], 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		_, err = db.Exec("DELETE FROM tils WHERE id = ?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/edit", func(w http.ResponseWriter, r *http.Request) {
		type JsonTil struct {
			ID string `json:"ID"`
			Title string `json:"Title"`
		}
		var til JsonTil

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&til)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err = db.Exec("UPDATE tils SET title = ? WHERE id = ?", til.Title, til.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	n := negroni.Classic()
	n.Use(sessions.Sessions("til", cookiestore.New([]byte("my-secret-123"))))
	n.Use(negroni.HandlerFunc(verifyDatabase))
	n.Use(negroni.HandlerFunc(verifyUser))
	n.UseHandler(mux)
	n.Run(":8080")
}