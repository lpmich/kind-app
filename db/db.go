package db

import (
    "os"
    "fmt"
    "time"
    "strconv"
    "database/sql"
    "github.com/google/uuid"
    "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Person struct {
    First string
    Last string
    Color string
}

type User struct {
    Username string
    Password string
    Salt []byte
}

type Post struct {
    Content string
    Author string
    Date time.Time
    Likes int
    Comments []Comment
    Id int
}

type Comment struct {
    Content string
    Author string
    Date time.Time
    Likes int
}

// Try to connect to database 10 times
func Conn() error {
    var err error
    for i:= 0; i < 10; i++ {
        err = initDB()
        if err != nil {
            time.Sleep(10 * time.Second)
            fmt.Println("Attempting to connect to database for", (i+1)*10, "seconds")
        } else { return nil }
    }
    return fmt.Errorf("Failed to connect to database after 10 tries", err)
}


// Initialize database
func initDB() error {
    cfg := mysql.Config{
        User: "root",
        Passwd: os.Getenv("MYSQL_ROOT_PASSWORD"),
        Net: "tcp",
        Addr: os.Getenv("MYSQL_URL")+":3306",
        ParseTime: true}
    var err error
    db, err = sql.Open("mysql", cfg.FormatDSN())
    if err != nil {
        return fmt.Errorf("Can't connect to database: ", err)
    }
    db.Exec("CREATE DATABASE kindapp")
    db.Exec("USE kindapp")
    people := `CREATE TABLE IF NOT EXISTS person(first VARCHAR(50) NOT NULL,
               last VARCHAR(50) NOT NULL, color VARCHAR(50) NOT NULL,
               id INTEGER AUTO_INCREMENT, PRIMARY KEY (id))`
    _, err = db.Exec(people)
    if err != nil {
        return fmt.Errorf("Error creating table people: ", err)
    }
    user := `CREATE TABLE IF NOT EXISTS user(username VARCHAR(50) NOT NULL,
             password CHAR(128) NOT NULL, id INTEGER AUTO_INCREMENT,
             salt BINARY(16) NOT NULL, PRIMARY KEY (id))`
    _, err = db.Exec(user)
    if err != nil {
        return fmt.Errorf("Error creating table user: ", err)
    }
    session := `CREATE TABLE IF NOT EXISTS session(uuid VARCHAR(50) NOT NULL,
                username VARCHAR(50) NOT NULL, PRIMARY KEY (uuid))`
    _, err = db.Exec(session)
    if err != nil {
        return fmt.Errorf("Error creating table session: ", err)
    }
    post := `CREATE TABLE IF NOT EXISTS post(content VARCHAR(1000) NOT NULL,
             author VARCHAR(50) NOT NULL, date DATETIME DEFAULT CURRENT_TIMESTAMP,
             likes INTEGER NOT NULL DEFAULT 0, id INTEGER AUTO_INCREMENT, PRIMARY KEY (id))`
    _, err = db.Exec(post)
    if err != nil {
        return fmt.Errorf("Error creating table post: ", err)
    }
    comment := `CREATE TABLE IF NOT EXISTS comment(content VARCHAR(500) NOT NULL,
                author VARCHAR(50) NOT NULL, date DATETIME DEFAULT CURRENT_TIMESTAMP,
                likes INTEGER NOT NULL DEFAULT 0, post_id INT NOT NULL,
                id INTEGER AUTO_INCREMENT,PRIMARY KEY (id),
                FOREIGN KEY (post_id) REFERENCES post(id) ON DELETE CASCADE ON UPDATE CASCADE)`
    _, err = db.Exec(comment)
    if err != nil {
        return fmt.Errorf("Error creating table comment: ", err)
    }

    fmt.Println("Database Connected!")
    return nil
}

// Returns username given a uuid 
func GetUsername(uuid string) (string, error) {
    var username string
    row, err := db.Query("SELECT username FROM session WHERE uuid='"+uuid+"'")
    if err != nil {
        return "", fmt.Errorf("Error retrieving username from session uuid: ", err)
    }
    defer row.Close()
    if row.Next() {
        err = row.Scan(&username)
        if err != nil {
            return "", fmt.Errorf("Error reading rows from session table", err)
        }
        return username, nil
    } else {
        return "", fmt.Errorf("Session does not exist")
    }
}

// Get user creds
func GetCreds(username string) (string, []byte, error) {
    var password string
    var hash []byte
    query := "SELECT password, salt FROM user WHERE username='" + username + "'"
    rows, err := db.Query(query)
    if err != nil {
        return "", nil, fmt.Errorf("Error retrieving from user table: ", err)
    }
    defer rows.Close()
    if rows.Next() {
        err = rows.Scan(&password, &hash)
        if err != nil {
            return "", nil, fmt.Errorf("Error reading rows from user table: ", err)
        }
    } else {
        return "", nil, fmt.Errorf("Username is incorrect.")
    }
    return password, hash, nil
}

// Gets people
func Getpeople()([]Person, error) {
    var people []Person

    rows, err := db.Query("SELECT first, last, color FROM person")
    if err != nil {
        return nil, fmt.Errorf("Error retrieving from person table: ", err)
    }
    defer rows.Close()
    for rows.Next() {
        var person Person
        err = rows.Scan(&person.First, &person.Last, &person.Color)
        if err != nil {
            return nil, fmt.Errorf("Error reading data: ", err)
        }
        people = append(people, person)
    }

    return people, nil
}

// Get comments for a given post
func GetComments(id int) ([]Comment, error) {
    var comments []Comment
    rows, err := db.Query("SELECT content, author, date, likes FROM comment WHERE post_id='"+
        strconv.Itoa(id)+"'")
    if err != nil {
        return nil, fmt.Errorf("Error retrieving from comment table: ", err)
    }
    defer rows.Close()
    for rows.Next() {
        var comment Comment
        err = rows.Scan(&comment.Content, &comment.Author, &comment.Date, &comment.Likes)
        if err != nil {
            return nil, fmt.Errorf("Error reading data: ", err)
        }
        comments = append(comments, comment)
    }
    return comments, nil
}

// Get all posts in the system
func GetAllPosts() ([]Post, error) {
    var posts []Post
    rows, err := db.Query("SELECT content, author, date, likes, id FROM post")
    if err != nil {
        return nil, fmt.Errorf("Error retrieving from post table: ", err)
    }
    defer rows.Close()
    for rows.Next() {
        var post Post
        err = rows.Scan(&post.Content, &post.Author, &post.Date, &post.Likes, &post.Id)
        if err != nil {
            return nil, fmt.Errorf("Error reading data: ", err)
        }
        posts = append(posts, post)
    }
    return posts, nil
}

// Retrieves a post with a given id
func GetPost(id string) (Post, error) {
    var post Post
    row, err := db.Query("SELECT content, author, date, likes, id FROM post WHERE id='"+id+"'")
    if err != nil {
        return post, fmt.Errorf("Error retrieving from post table: ", err)
    }
    defer row.Close()
    if row.Next() {
        err = row.Scan(&post.Content, &post.Author, &post.Date, &post.Likes, &post.Id)
        if err != nil {
            return post, fmt.Errorf("Error reading data: ", err)
        }
    } else {
        return post, fmt.Errorf("Post "+id+" does not exist.")
    }
    return post, nil
}

// Adds a person
func Addperson(person Person) error {
    _, err := db.Exec("INSERT INTO person (first, last, color) VALUES (?, ?, ?)",
        person.First, person.Last, person.Color)
    if err != nil {
        return fmt.Errorf("Error inserting into person table: ", err)
    }
    return nil
}

// Adds a user
func Adduser(user User) error {
    _, err := db.Exec("INSERT INTO user (username, password, salt) VALUES (?, ?, ?)",
        user.Username, user.Password, user.Salt)
    if err != nil {
        return fmt.Errorf("Error inserting into user table: ", err)
    }
    return nil
}

// Adds a user's session
func AddSession(username string) (string, error) {
    err := DeleteSession(username)
    if err != nil {
        return "", err
    }
    var id string
    for ; true; {
        id = uuid.NewString()
        row, _ := db.Query("SELECT * FROM session where uuid='"+id+"'")
        if row.Next() {
            // Generate new id
        } else {
            break
        }
    }
    _, err = db.Exec("INSERT INTO session (uuid, username) VALUES (?, ?)", id, username)
    if err != nil {
        return "", fmt.Errorf("Error inserting into session table: ", err)
    }
    return id, nil
}

// Deletes a user's session
func DeleteSession(username string) error {
    _, err := db.Exec("DELETE FROM session WHERE username='"+username+"'")
    if err != nil {
        return fmt.Errorf("Error removing previous session for user "+username+": ", err)
    }
    return nil
}

// Determines if a session id is valid or not
func ValidSession(uuid string) (bool, error) {
    row, err := db.Query("SELECT uuid FROM session WHERE uuid='"+uuid+"'")
    if err != nil {
        return false, fmt.Errorf("Error retrieving session: ", err)
    }
    defer row.Close()
    if row.Next() {
        var id string
        err = row.Scan(&id)
        if err != nil {
            return false, fmt.Errorf("Error reading from session rows: ", err)
        }
        if uuid == id {
            return true, nil
        }
    }
    return false, nil
}

// Gets the author of a post or comment
func GetAuthor(entity string, id int) (string, error) {
    var author string
    row, err := db.Query("SELECT author FROM "+entity+" WHERE id='"+strconv.Itoa(id)+"'")
    if err != nil {
        return "", err
    }
    defer row.Close()
    if row.Next() {
        err = row.Scan(&author)
        if err != nil {
            return "", fmt.Errorf("Error reading from author rows: ", err)
        }
    } else {
        return "", fmt.Errorf(entity+" with id:"+strconv.Itoa(id)+" does not exist")
    }
    return author, nil
}

func GetPostIDFromCommentID(commentID int) (int, error) {
    var postID int
    row, err := db.Query("SELECT post_id FROM comment WHERE id='"+strconv.Itoa(commentID)+"'")
    if err != nil {
        return 0, err
    }
    defer row.Close()
    if row.Next() {
        err = row.Scan(&postID)
        if err != nil {
            return 0, fmt.Errorf("Error reading from comment rows: ", err)
        }
    } else {
        return 0, fmt.Errorf("Comment "+strconv.Itoa(commentID)+" cannot be linked to a post")
    }
    return postID, nil
}

// Adds a post
func AddPost(content string, author string) error {
    _, err := db.Exec("INSERT INTO post (content, author) VALUES (?, ?)", content, author)
    return err
}

// Deletes a post
func DeletePost(id int) error {
    _, err := db.Exec("DELETE FROM post WHERE id='"+strconv.Itoa(id)+"'")
    return err
}

// Adds a comment to a post
func AddComment(content string, author string, post_id int) error {
    _, err := db.Exec("INSERT INTO comment (content, author, post_id) VALUES (?, ?, ?)",
        content, author, post_id)
    return err
}

// Deletes a comment from a post
func DeleteComment(id int) error {
    _, err := db.Exec("DELETE FROM comment WHERE id='"+strconv.Itoa(id)+"'")
    return err
}

// Returns the number of likes associate with a post or comment
func GetLikes(entity string, id int) (int, error) {
    var num_likes int
    row, err := db.Query("SELECT likes FROM "+entity+" WHERE id='"+strconv.Itoa(id)+"'")
    if err != nil {
        return 0, err
    }
    defer row.Close()
    if row.Next() {
        err = row.Scan(&num_likes)
        if err != nil {
            return 0, fmt.Errorf("Error reading from "+entity+" row: ", err)
        }
    } else {
        return 0, fmt.Errorf(entity+" with id:"+strconv.Itoa(id)+" not found")
    }
    return num_likes, nil
}

// Likes a post or comment
func Like(entity string, id int) error {
    num_likes, err := GetLikes(entity, id)
    num_likes++
    if err != nil {
        return err
    }
    _, err = db.Exec("UPDATE "+entity+" SET likes="+strconv.Itoa(num_likes)+
        " WHERE id='"+strconv.Itoa(id)+"'")
    return err
}

// Dislikes a post or comment
func Dislike(entity string, id int) error {
    num_likes, err := GetLikes(entity, id)
    if err != nil {
        return err
    }
    if num_likes > 0 {
        num_likes--
    }
    _, err = db.Exec("UPDATE "+entity+" SET likes="+strconv.Itoa(num_likes)+
        " WHERE id='"+strconv.Itoa(id)+"'")
    return err
}
