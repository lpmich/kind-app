package main

import (
    "log"
    "fmt"
    "net/http"
    "text/template"
    "github.com/lpmich/kind-app/db"
    "github.com/lpmich/kind-app/security"
)

type HTMLData struct {
    People []db.Person
    Username string
}

type HTTPError struct {
    Message string
}

// Redirects http to https
func redirectHTTP(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "https://localhost"+r.RequestURI, 302)
}

// Retrieve session uuid from cookies
func getSessionID(r *http.Request) string {
    cookie, err := r.Cookie("sessionid")
    if err != nil { // Cookie doesn't exist
        return ""
    }
    return cookie.Value
}

// Checks for cookie to see if authenticated, otherwise directs to login
func isAuthenticated(r *http.Request) bool {
    uuid := getSessionID(r)
    if uuid == "" {
        return false
    }
    valid, err := security.IsAuthenticated(uuid)
    if err != nil {
        fmt.Println("Error validating session:", err)
        return false
    }
    return valid
}

// Serve index.html
func index(w http.ResponseWriter, r *http.Request) {
    if !isAuthenticated(r) {
        http.Redirect(w, r, "https://localhost/login", 303)
        return
    }

    if r.Method == "GET" {
        t, _ := template.ParseFiles("assets/index.html")
        t.Execute(w, nil)
    }

    if r.Method == "POST" {
        fname := r.FormValue("first")
        lname := r.FormValue("last")
        color := r.FormValue("color")

        // Add person to database
        var person db.Person
        person.First = fname
        person.Last = lname
        person.Color = color
        db.Addperson(person)
        http.Redirect(w, r, "https://localhost/view", 303)
    }
}

// Serve view.html
func view(w http.ResponseWriter, r *http.Request) {
    var data HTMLData
    if !isAuthenticated(r) {
        http.Redirect(w, r, "https://localhost/login", 303)
        return
    }
    people, _ := db.Getpeople()
    data.People = people
    t, _ := template.ParseFiles("assets/view.html")
    t.Execute(w, data)
}

// Creates a user
func createUser(w http.ResponseWriter, r *http.Request) {
    if isAuthenticated(r) {
        redirectHTTP(w, r)
    }
    if r.Method == "GET" {
        t, _ := template.ParseFiles("assets/createuser.html")
        t.Execute(w, nil)
    }
    if r.Method == "POST" {
        username := r.FormValue("username")
        password := r.FormValue("password")
        err := security.Createuser(username, password)
        if err != nil {
            fmt.Println(err)
            httpError := HTTPError{
                Message: err.Error(),
            }
            t, _ := template.ParseFiles("assets/createuser.html")
            t.Execute(w, httpError)
        } else {
            http.Redirect(w, r, "https://localhost/login", 303)
        }
    }
}

// Authenticates a user
func login(w http.ResponseWriter, r *http.Request) {
    if isAuthenticated(r) {
        http.Redirect(w, r, "https://localhost/", 303)
    }
    if r.Method == "GET" {
        t, _ := template.ParseFiles("assets/login.html")
        t.Execute(w, nil)
    }
    if r.Method == "POST" {
        username := r.FormValue("username")
        password := r.FormValue("password")
        uuid, err := security.Authenticate(username, password)
        if err != nil {
            fmt.Println(err)
            httpError := HTTPError{
                Message: err.Error(),
            }
            t, _ := template.ParseFiles("assets/login.html")
            t.Execute(w, httpError)
        } else {
            // Add session cookie
            c := &http.Cookie{
                Name: "sessionid",
                Value: uuid,
                MaxAge: 0,
            }
            http.SetCookie(w, c)
            http.Redirect(w, r, "https://localhost", 303)
        }
    }
}

func logout(w http.ResponseWriter, r *http.Request) {
    if !isAuthenticated(r) {
        http.Redirect(w, r, "https://localhost/login", 303)
    }
    uuid := getSessionID(r)
    err := security.RemoveSession(uuid)
    if err != nil {
        fmt.Println(err)
    }
    http.Redirect(w, r, "https://localhost/login", 303)
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
    http.HandleFunc("/", index)
    http.HandleFunc("/view", view)
    http.HandleFunc("/login", login)
    http.HandleFunc("/logout", logout)
    http.HandleFunc("/createuser", createUser)
    go http.ListenAndServe(":80", http.HandlerFunc(redirectHTTP))
    log.Fatal(http.ListenAndServeTLS(":443", "security/server.pem", "security/server.key", nil))
}
