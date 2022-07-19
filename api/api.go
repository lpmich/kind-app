package api

import (
    "fmt"
    "time"
    "net/http"
    "strings"
    "encoding/json"
    "gitlab.sas.com/lomich/kind-app/db"
    "gitlab.sas.com/lomich/kind-app/security"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt"
)

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

var signingKey = []byte("pvh7YQqceEC0qqKUSCTh")

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

    m := make(map[string]string)
    m["key"] = tokenString
    jsonString, _ := json.Marshal(m)

    c.IndentedJSON(http.StatusOK, jsonString)
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
            fmt.Println()
            if err == jwt.ErrSignatureInvalid {
                c.JSON(http.StatusUnauthorized, gin.H{
                    "error": err.Error()})
                return false
            }
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return false
        }
        if !tkn.Valid {
            fmt.Println()
            c.JSON(http.StatusUnauthorized, gin.H{"error":
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

// Landing page for API
func apiLanding(c *gin.Context) {
    if !authorized(c) {
        return
    }
    c.String(http.StatusOK, "Welcome to kind-app API")
}

// Gets a post by id
func getPost(c *gin.Context) {
    if !authorized(c) {
        return
    }
    var post post

    id := c.Param("id")
    db_post, err := db.GetPost(id)
    if err != nil {
        fmt.Println(err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    post.Content = db_post.Content
    post.Author = db_post.Author
    post.Date = db_post.Date.String()
    post.Likes = db_post.Likes
    post.Comments = fmt.Sprintf("%#v", db_post.Comments)
    post.Id = db_post.Id
    c.IndentedJSON(http.StatusOK, post)
}

// Get all posts in the system
func getPosts(c *gin.Context) {
    if !authorized(c) {
        return
    }
    var posts []post
    db_posts, err := db.GetAllPosts()
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    for _, db_post := range db_posts {
        var post post
        post.Content = db_post.Content
        post.Author = db_post.Author
        post.Date = db_post.Date.String()
        post.Likes = db_post.Likes
        post.Comments = fmt.Sprintf("%#v", db_post.Comments)
        post.Id = db_post.Id
        posts = append(posts, post)
    }
    c.IndentedJSON(http.StatusOK, posts)
}

// Creates a post with content and author
func postPost(c *gin.Context) {
    c.IndentedJSON(http.StatusOK, nil)
}

// Get all comments in the system
func getComments(c *gin.Context) {
    c.IndentedJSON(http.StatusOK, nil)

}

// Initialize GIN API and expose endpoints
func StartAPI() {
    router := gin.Default()

    router.GET("", apiLanding)
    router.POST("/api/jwt", generateJWT)

    router.GET("/api/posts", getPosts)
    router.GET("/api/post/:id", getPost)
    router.POST("/api/post", postPost)

    router.GET("/api/comments", getComments)
    router.RunTLS(":8080", "security/server.crt", "security/server.key")
}

