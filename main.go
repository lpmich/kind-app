package main

import (
    "log"
    "fmt"
    "strconv"
    "net/http"
    "text/template"
    "github.com/lpmich/kind-app/db"
    "github.com/lpmich/kind-app/security"
)

type HTMLData struct {
    People []db.Person
    Username string
}

type IndexData struct {
    Posts []db.Post
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

    var data IndexData
    posts, err := db.GetAllPosts()
    if err != nil {
        fmt.Println(err)
        return
    }
    for _, post := range posts {
        comments, _ := db.GetComments(post.Id)
        post.Comments = comments
    }
    data.Posts = posts
    t, _ := template.ParseFiles("assets/index.html")
    t.Execute(w, data)
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

// Logs a user out of their current session
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

// Creates a new post
func post(w http.ResponseWriter, r *http.Request) {
    if !isAuthenticated(r) {
        http.Redirect(w, r, "https://localhost/login", 303)
    }
    uuid := getSessionID(r)
    author, err := db.GetUsername(uuid)
    if err != nil {
        fmt.Println(err)
        return
    }
    content := r.FormValue("content")
    err = db.AddPost(content, author)
    if err != nil {
        fmt.Println(err)
    }
    http.Redirect(w, r, "https://localhost/", 303)
}

// Removes a post
func removePost(w http.ResponseWriter, r *http.Request) {
    if !isAuthenticated(r) {
        http.Redirect(w, r, "https://localhost/login", 303)
    }
    uuid := getSessionID(r)
    username, err := db.GetUsername(uuid)
    if err != nil {
        fmt.Println(err)
        return
    }
    id, _ := strconv.Atoi(r.FormValue("postid"))
    author, err := db.GetAuthor("post", id)
    if author != username {
        fmt.Println("User: "+username+" is not authorized to delete "+author+"'s post")
        return
    }
    err = db.DeletePost(id)
    if err != nil {
        fmt.Println(err)
    }
}

// Creates a new comment
func comment(w http.ResponseWriter, r *http.Request) {
    if !isAuthenticated(r) {
        http.Redirect(w, r, "https://localhost/login", 303)
    }
    uuid := getSessionID(r)
    author, err := db.GetUsername(uuid)
    if err != nil {
        fmt.Println(err)
        return
    }
    id, _ := strconv.Atoi(r.FormValue("postid"))
    content := r.FormValue("content")
    err = db.AddComment(content, author, id)
    if err != nil {
        fmt.Println(err)
    }
}

// Removes a comment
func removeComment(w http.ResponseWriter, r *http.Request) {
    if !isAuthenticated(r) {
        http.Redirect(w, r, "https://localhost/login", 303)
    }
    uuid := getSessionID(r)
    username, err := db.GetUsername(uuid)
    if err != nil {
        fmt.Println(err)
        return
    }
    id, _ := strconv.Atoi(r.FormValue("commentid"))
    commentAuthor, err := db.GetAuthor("comment", id)
    if err != nil {
        fmt.Println(err)
        return
    }
    postID, err := db.GetPostIDFromCommentID(id)
    if err != nil {
        fmt.Println(err)
        return
    }
    postAuthor, err := db.GetAuthor("post", postID)
    if err != nil {
        fmt.Println(err)
        return
    }
    if commentAuthor != username || postAuthor != username {
        fmt.Println("User: "+username+" is not authorized to delete "+commentAuthor+"'s comment")
        return
    }
    err = db.DeleteComment(id)
    if err != nil {
        fmt.Println(err)
    }
}

// Likes a post/comment
func like(w http.ResponseWriter, r *http.Request) {
    if !isAuthenticated(r) {
        http.Redirect(w, r, "https://localhost/login", 303)
    }
    entity := r.FormValue("entity")
    id, _ := strconv.Atoi(r.FormValue("id"))
    err := db.Like(entity, id)
    if err != nil {
        fmt.Println(err)
    }
}

// Dislikes a post/comment
func dislike(w http.ResponseWriter, r *http.Request) {
    if !isAuthenticated(r) {
        http.Redirect(w, r, "https://localhost/login", 303)
    }
    entity := r.FormValue("entity")
    id, _ := strconv.Atoi(r.FormValue("id"))
    err := db.Dislike(entity, id)
    if err != nil {
        fmt.Println(err)
    }
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
    http.HandleFunc("/createuser", createUser)
    http.HandleFunc("/login", login)
    http.HandleFunc("/logout", logout)
    http.HandleFunc("/post", post)
    http.HandleFunc("/removepost", removePost)
    http.HandleFunc("/comment", comment)
    http.HandleFunc("/removecomment", removeComment)
    http.HandleFunc("/like", like)
    http.HandleFunc("/dislike", dislike)
    http.HandleFunc("/view", view)
    go http.ListenAndServe(":80", http.HandlerFunc(redirectHTTP))
    log.Fatal(http.ListenAndServeTLS(":443", "security/server.crt", "security/server.key", nil))
}
