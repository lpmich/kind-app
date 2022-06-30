package db

import (
    "os"
    "fmt"
    "time"
    "database/sql"
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

    fmt.Println("Database Connected!")
    return nil
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
    message := "Error reading rows from user table: "
    if rows.Next() {
        err := rows.Scan(&password, &hash)
        if err != nil {
            return "", hash, fmt.Errorf(message, err)
        }
    } else {
        return "", nil, fmt.Errorf(message, err)
    }
    return password, hash, nil
}


// Get people
func Getpeople()([]Person, error) {
    var people []Person

    rows, err := db.Query("SELECT first, last, color FROM person")
    if err != nil {
        return nil, fmt.Errorf("Error retrieving from person table: ", err)
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
func Addperson(person Person) error {
    _, err := db.Exec("INSERT INTO person (first, last, color) VALUES (?, ?, ?)",
        person.First, person.Last, person.Color)
    if err != nil {
        return fmt.Errorf("Error inserting into person table: ", err)
    }
    return nil
}

func Adduser(user User) error {
    _, err := db.Exec("INSERT INTO user (username, password, salt) VALUES (?, ?, ?)",
        user.Username, user.Password, user.Salt)
    if err != nil {
        return fmt.Errorf("Error inserting into user table: ", err)
    }
    return nil
}
