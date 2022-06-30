package main

import (
    "log"
    "fmt"
    "net/http"
    "text/template"
    "database/sql"
    "github.com/go-sql-driver/mysql"
)

type Person struct {
    First string
    Last string
    Color string
}

// Redirects http to https
func redirectHTTP(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "https://localhost"+r.RequestURI, 302)
}

// Serve index.html
func indexHandler(w http.ResponseWriter, r *http.Request) {
    http.ServeFile( w, r, "assets/index.html")
}

// Serve view.html
func viewHandler(w http.ResponseWriter, r *http.Request) {
    people, _ := db.Getpeople()
    t, _ := template.ParseFiles("assets/view.html")
    t.Execute(w, people)
}

// Process input from form
func processHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Redirect(w, r, "/", http.StatusSeeOther)
        return
    }

    fname := r.FormValue("first")
    lname := r.FormValue("last")
    color := r.FormValue("color")

    // Add person to database
    var person db.Person
    person.First = fname
    person.Last = lname
    person.Color = color
    db.Addperson(person)
    http.Redirect(w, r, "/view", 303)
}

// Serve view.html
func viewHandler(w http.ResponseWriter, r *http.Request) {
    people, _ := getPeople()
    t, _ := template.ParseFiles("assets/view.html")
    t.Execute(w, people)
}

// Redirects http to https
func redirectHTTP(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "https://localhost"+r.RequestURI, 302)
}

// Serve application
func main() {

    fmt.Println("Starting Application...")

    // Connect to database
    err := db.Conn()
    if err != nil {
        log.Fatal(err)
    }

    // Listen for http/s requests
    fmt.Println("Serving Application...")
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/process", processHandler)
    http.HandleFunc("/view", viewHandler)
    errs := make(chan error, 1)
    go http.ListenAndServe(":80", http.HandlerFunc(redirectHTTP))
    go http.ListenAndServeTLS(":443", "security/server.pem", "security/server.key", nil)
    log.Fatal(<-errs)
}
