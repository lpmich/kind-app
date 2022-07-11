package security

import (
    "fmt"
    "github.com/lpmich/kind-app/db"
    "crypto/rand"
    "crypto/sha512"
    "encoding/base64"
)

// Hashes a salted password with SHA512 and returns encoded result
func hashPassword(password string, salt []byte) string {
    hasher := sha512.New()
    passwordBytes := []byte(password)
    passwordBytes = append(passwordBytes, salt...)
    hasher.Write(passwordBytes)
    hash := hasher.Sum(nil)
    return base64.URLEncoding.EncodeToString(hash)
}

// Authenticates a user's credentials
func Authenticate(username string, password string) (string, error) {
    hash, salt, err := db.GetCreds(username)
    if err != nil {
        return "", err
    }
    hashedPassword := hashPassword(password, salt)

    if hashedPassword != hash {
        return "", fmt.Errorf("Password is incorrect")
    }
    uuid, err := db.AddSession(username)
    if err != nil {
        return "", fmt.Errorf("Error creating session. ", err)
    }
    return uuid, nil
}

// Creates a new user
func Createuser(username string, password string) error {
    hash, salt, _ := db.GetCreds(username)
    if hash != "" {
        return fmt.Errorf("User already exists")
    }

    salt = make([]byte, 16)
    _, err := rand.Read(salt)
    if err != nil {
        return fmt.Errorf("Error creating salt: ", err)
    }

    hash = hashPassword(password, salt)
    user := db.User{username, hash, salt}
    return db.Adduser(user)
}

// Determines if a user is authenticated or not
func IsAuthenticated(uuid string) (bool, error) {
    return db.ValidSession(uuid)
}

// Removes a session given a uuid
func RemoveSession(uuid string) error {
    username, err := db.GetUsername(uuid)
    if err != nil {
        return err
    }
    return db.DeleteSession(username)
}
