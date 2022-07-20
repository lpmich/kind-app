package api

import (
    "fmt"
    "time"
    "net/http"
    "strings"
    "encoding/json"
    "gitlab.sas.com/lomich/kind-app/db"
    "gitlab.sas.com/lomich/kind-app/security"
    "github.com/google/uuid"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt"
)

// local objects to read-in & output json body
type post struct {
    Content string `json:"content"`
    Author string `json:"author"`
    Date string `json:"date"`
    Likes int `json:"likes"`
    Comments string `json:"comments"`
    Id  string `json:"id"`
}

type credentials struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type claims struct {
    Username string `json:"username"`
    jwt.StandardClaims
}

type newContent struct {
    Content string `json:"content"`
}


var signingKey = []byte(uuid.NewString())

// Generates a new JWT
func generateJWT(c *gin.Context) {
    var creds credentials

    // Decode body and read it into a credentials object
    err := json.NewDecoder(c.Request.Body).Decode(&creds)
    if err != nil {
        fmt.Println(err)
        c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Check user is verified
    _, err = security.Authenticate(creds.Username, creds.Password)
    if err != nil {
        fmt.Println(err)
        c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Make JWT
    expirationTime := time.Now().Add(time.Hour * 168)
    claim := &claims {
        Username: creds.Username,
        StandardClaims: jwt.StandardClaims {
            ExpiresAt: expirationTime.Unix()},
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS512, claim)
    tokenString, err := token.SignedString(signingKey)
    if err != nil {
        fmt.Println(err)
        c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.IndentedJSON(http.StatusOK, gin.H{"key": tokenString})
}

// Determines if a request is authorized
func authorized(c *gin.Context) bool {
    authHeader := c.Request.Header["Authorization"]
    if len(authHeader) > 0 {
        jwtString := strings.Fields(authHeader[0])[1]
        claims := &claims{}
        tkn, err := jwt.ParseWithClaims(jwtString, claims,
            func(t *jwt.Token) (interface{}, error) {
                return signingKey, nil
            })
        if err != nil {
            c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error parsing JWT Claims: "+err.Error()})
            return false
        }
        if !tkn.Valid {
            fmt.Println()
            c.IndentedJSON(http.StatusUnauthorized, gin.H{"error":
                "jwt is not valid"})
            return false
        }
        return true
    } else {
        cookie, _ := c.Request.Cookie("sessionid")
        if cookie == nil {
            return false
        }
        uuid := cookie.Value
        authorized, _ := security.IsAuthenticated(uuid)
        return authorized
    }
    return false
}

// Return username from encoded JWT
func getUsername(c *gin.Context) string {
    authHeader := c.Request.Header["Authorization"]
    jwtString := strings.Fields(authHeader[0])[1]
    claims := &claims{}
    jwt.ParseWithClaims(jwtString, claims,
        func(t *jwt.Token) (interface{}, error) {
            return signingKey, nil
    })
    return claims.Username
}

// Landing page for API
func apiLanding(c *gin.Context) {
    if !authorized(c) {
        c.IndentedJSON(http.StatusUnauthorized, gin.H{"error":
            "You are not authorized, ensure your JWT is presented correctly"})
        return
    }
    c.String(http.StatusOK, "Welcome to kind-app API")
}

// Gets a post by id
func getPost(c *gin.Context) {
    if !authorized(c) {
        c.IndentedJSON(http.StatusUnauthorized, gin.H{"error":
            "You are not authorized, ensure your JWT is presented correctly"})
        return
    }
    var post post

    id := c.Param("id")
    db_post, err := db.GetPost(id)
    if err != nil {
        fmt.Println(err)
        c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    post.Content = db_post.Content
    post.Author = db_post.Author
    post.Date = db_post.Date.String()
    post.Likes = db_post.Likes
    post.Id = db_post.Id

    comments, _ := db.GetComments(post.Id)
    post.Comments = fmt.Sprintf("%#v", comments)
    c.IndentedJSON(http.StatusOK, post)
}

// Get all posts in the system
func getPosts(c *gin.Context) {
    if !authorized(c) {
        c.IndentedJSON(http.StatusUnauthorized, gin.H{"error":
            "You are not authorized, ensure your JWT is presented correctly"})
        return
    }
    var posts []post
    db_posts, err := db.GetAllPosts()
    if err != nil {
        c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    for _, db_post := range db_posts {
        var post post
        post.Content = db_post.Content
        post.Author = db_post.Author
        post.Date = db_post.Date.String()
        post.Likes = db_post.Likes
        post.Id = db_post.Id

        comments, _ := db.GetComments(post.Id)
        post.Comments = fmt.Sprintf("%#v", comments)
        posts = append(posts, post)
    }
    c.IndentedJSON(http.StatusOK, posts)
}

// Creates a post with content and author
func postPost(c *gin.Context) {
    if !authorized(c) {
        c.IndentedJSON(http.StatusUnauthorized, gin.H{"error":
            "You are not authorized, ensure your JWT is presented correctly"})
        return
    }
    var p newContent
    err := json.NewDecoder(c.Request.Body).Decode(&p)
    if err != nil {
        c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error reading json body: "+err.Error()})
        return
    }
    username := getUsername(c)
    if strings.TrimSpace(p.Content) == "" {
        c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Content field cannot be empty"})
    }
    id, err := db.AddPost(p.Content, username)
    if err != nil {
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.IndentedJSON(http.StatusOK, gin.H{"message": "Success", "post_id": id})
}

func postComment(c *gin.Context) {
    if !authorized(c) {
        c.IndentedJSON(http.StatusUnauthorized, gin.H{"error":
            "You are not authorized, ensure your JWT is presented correctly"})
        return
    }

    username := getUsername(c)
    id := c.Param("id")

    var newComment newContent
    err := json.NewDecoder(c.Request.Body).Decode(&newComment)
    if err != nil {
        c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Error reading json body: "+err.Error()})
        return
    }
    if strings.TrimSpace(newComment.Content) == "" {
        c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Content field cannot be empty"})
        return
    }

    id, err = db.AddComment(newComment.Content, username, id)
    if err != nil {
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.IndentedJSON(http.StatusOK, gin.H{"message": "Success", "comment_id": id})
}


// Deletes a post
func deletePost(c *gin.Context) {
    if !authorized(c) {
        c.IndentedJSON(http.StatusUnauthorized, gin.H{"error":
            "You are not authorized, ensure your JWT is presented correctly"})
        return
    }
    username := getUsername(c)
    id := c.Param("id")

    author, err := db.GetAuthor("post", id)
    if err != nil {
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if username != author {
        c.IndentedJSON(http.StatusUnauthorized, gin.H{"error":
            "You do not have permission to delete a post that is not yours"})
        return
    }
    err = db.DeletePost(id)
    if err != nil {
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.IndentedJSON(http.StatusOK, gin.H{"message": "Success"})
}

// Deletes a comment
func deleteComment(c *gin.Context) {
    if !authorized(c) {
        c.IndentedJSON(http.StatusUnauthorized, gin.H{"error":
            "You are not authorized, ensure your JWT is presented correctly"})
        return
    }
    username := getUsername(c)
    id := c.Param("id")

    postID, err := db.GetPostIDFromCommentID(id)
    if err != nil {
        fmt.Println(err)
        return
    }
    postAuthor, err := db.GetAuthor("post", postID)
    if err != nil {
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    commentAuthor, err := db.GetAuthor("comment", id)
    if err != nil {
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if postAuthor != username && commentAuthor != username {
        c.IndentedJSON(http.StatusUnauthorized, gin.H{"error":
            "User: "+username+" is not authorized to delete "+commentAuthor+"'s comment"})
        return
    }
    err = db.DeleteComment(id)
    if err != nil {
        c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.IndentedJSON(http.StatusOK, gin.H{"message": "Success"})
}

// Initialize GIN API and expose endpoints
func StartAPI() {
    router := gin.Default()

    router.GET("", apiLanding)
    router.POST("/api/jwt", generateJWT)

    router.GET("/api/posts", getPosts)
    router.GET("/api/post/:id", getPost)

    router.POST("/api/post", postPost)
    router.POST("/api/comment/:id", postComment)

    router.DELETE("/api/post/:id", deletePost)
    router.DELETE("/api/comment/:id", deleteComment)

    router.RunTLS(":8080", "security/server.crt", "security/server.key")
}

