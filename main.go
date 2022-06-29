package main

import (
    "os"
    "log"
    "fmt"
    "time"
    "net/http"
    "text/template"
    "database/sql"
    "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Person struct {
    First string
    Last string
    Color string
}

// Create database
func createDB() error {
    cfg := mysql.Config{
        User: "root",
        Passwd: os.Getenv("MYSQL_ROOT_PASSWORD"),
        Net: "tcp",
        Addr: os.Getenv("MYSQL_URL")+":3306"}
    var err error
    db, err = sql.Open("mysql", cfg.FormatDSN())
    if err != nil {
        log.Fatal("Can't connect to database: %v", err)
    }
    db.Exec("CREATE DATABASE people;")
    db.Exec("USE people")
    query := `CREATE TABLE IF NOT EXISTS people(first VARCHAR(50) NOT NULL,
              last VARCHAR(50) NOT NULL, color VARCHAR(50) NOT NULL,
              id INTEGER AUTO_INCREMENT, PRIMARY KEY (id))`
    _, err = db.Exec(query)
    if err != nil {
        return err
    }
    fmt.Println("Database Connected!")
    return nil
}

// Retrieve all people in database
func getPeople()([]Person, error) {
    var people []Person

    rows, err := db.Query("SELECT first, last, color FROM people")
    if err != nil {
        return nil, fmt.Errorf("Error retrieving from database: ", err)
    }
    defer rows.Close()
    for rows.Next() {
        var person Person
        if err := rows.Scan(&person.First, &person.Last, &person.Color); err != nil {
            return nil, fmt.Errorf("Error reading data: ", err)
        }
        people = append(people, person)
    }

    return people, nil
}

// Add a person to database
func addPerson(person Person) {
    _, err := db.Exec("INSERT INTO people (first, last, color) VALUES (?, ?, ?)",
        person.First, person.Last, person.Color)
    if err != nil {
        log.Fatal("Error inserting into database: %v", err)
    }
}

// Serve index.html
func indexHandler(w http.ResponseWriter, r *http.Request) {
    http.ServeFile( w, r, "assets/index.html")
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
    var person Person
    person.First = fname
    person.Last = lname
    person.Color = color
    addPerson(person)
    http.Redirect(w, r, "/view", 303)
}

// Serve view.html
func viewHandler(w http.ResponseWriter, r *http.Request) {
    people, _ := getPeople()
    t, _ := template.ParseFiles("assets/view.html")
    t.Execute(w, people)
}

// Serve application
func main() {
    fmt.Println("Starting Application...")

    // Try to connect to database 10 times
    var err error
    for i:= 0; i < 10; i++ {
        if i > 0 {
           log.Println("Attempting to connect to Database: ", i)
           time.Sleep(10 * time.Second)
        }
        err = createDB()
        if err == nil {
            break
        }
    }
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Serving application")
    // Listen and handle http requests
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/process", processHandler)
    http.HandleFunc("/view", viewHandler)
    http.ListenAndServe(":8080", nil)
}
