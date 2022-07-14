package api

import (
    "fmt"
    "net/http"
    "gitlab.sas.com/lomich/kind-app/db"
    "gitlab.sas.com/lomich/kind-app/security"
    "github.com/gin-gonic/gin"
//    "github.com/golang-jwt/jwt"
)

type post struct {
    Content string `json:"content"`
    Author string `json:"author"`
    Date string `json:"date"`
    Likes int `json:"likes"`
    Comments string `json:"comments"`
    Id  int `json:"id"`
}

// Initialize GIN API and expose endpoints
func StartAPI() {
    router := gin.Default()
    router.GET("", apiLanding)
    router.GET("/api/post/:id", getPost)
    router.GET("/api/posts", getPosts)
    router.POST("api/post", postPost)
    router.GET("/api/comments", getComments)
    router.RunTLS(":8080", "security/server.crt", "security/server.key")
}

// Determines if a request is authorized
func authorized(c *gin.Context) bool {
    authHeader := c.Request.Header["Authorization"]
    if len(authHeader) > 0 {
//        jwt := strings.Fields(authHeader[0])[1]
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
    authorized(c)
    c.String(http.StatusOK, "Welcome to kind-app API")
}

// Gets a post by id
func getPost(c *gin.Context) {
    var post post

    id := c.Param("id")
    db_post, err := db.GetPost(id)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
