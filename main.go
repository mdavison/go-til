package main

import (
	"net/http"
	"html/template"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"

	"encoding/json"

	"github.com/urfave/negroni"
	gmux "github.com/gorilla/mux"
	"strconv"
)

type Page struct {
	Tils []Til
}

type Til struct {
	ID int
	Title string
}

var db *sql.DB

func verifyDatabase(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if err := db.Ping(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	next(w, r)
}

func main() {
	templates := template.Must(template.ParseFiles("templates/index.html"))

	db, _ = sql.Open("sqlite3", "til.db")

	mux := gmux.NewRouter()

	s := http.StripPrefix("/resources/", http.FileServer(http.Dir("./resources/")))
	mux.PathPrefix("/resources/").Handler(s)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
		p := Page{ Tils: []Til{} }
		rows, _ := db.Query("SELECT id, title FROM tils")
		for rows.Next() {
			var t Til
			rows.Scan(&t.ID, &t.Title)
			p.Tils = append(p.Tils, t)
		}

		if err := templates.ExecuteTemplate(w, "index.html", p); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request){
		title := r.FormValue("title")

		row, err := db.Exec("INSERT INTO tils (id, title) values (?, ?)", nil, title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		id, _ := row.LastInsertId()
		results := []Til {
			Til{ID: int(id), Title: title},
		}

		encoder := json.NewEncoder(w)
		if err = encoder.Encode(results); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
	n.Use(negroni.HandlerFunc(verifyDatabase))
	n.UseHandler(mux)
	n.Run(":8080")
}