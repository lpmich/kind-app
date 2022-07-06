package db

import (
    "os"
    "fmt"
    "time"
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
        Addr: os.Getenv("MYSQL_URL")+":3306"}
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
func Getcreds(username string) (string, []byte, error) {
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
