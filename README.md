# Kind App
A simple golang web application deployed through [Kind Kubernetes](https://kind.sigs.k8s.io/)

## App
### Prerequisites
- [WSL2 or linux shell](https://docs.microsoft.com/en-us/windows/wsl/install)
- [Docker Engine](https://docs.docker.com/engine/install/)
- [Go](https://gist.github.com/alexchiri/aca79caee89a33f0856951cedbf306dc#install-go)
- [Kind](https://gist.github.com/alexchiri/aca79caee89a33f0856951cedbf306dc#install-kind)

### Installation
```bash
git pull https://gitlab.sas.com/lomich/kind-app.git
cd kind-app/
```

### Configuration
```bash
vi config/mysql-secret.yml
```
Insert the following code into the config/mysql-secret.yml, enter a desired password for the mysql database into the password field and then save the file with :wq
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mysql-secret
type: kubernetes.io/basic-auth
stringData:
  password:
```

### Run the Application
```
./build.sh
```
After the script finishes the application should be running at https://localhost

## About
This application is a basic blog where users can login, post, and comment.

### Core Models
```
user {
        username: string
        password: string
}
```
```
post {
        content:     string
        author:      string
        date:        date
        likes:       int
        comments:    []comment
        id:          int
}
```
```
comment {
        content:    string
        author:     string
        date:       date
        like:       int
        post_id:    int
        id:         int
}
```
## API
An API is accessible running on https://localhost:8080. Users must authenticate to the API through a [JWT](https://jwt.io). In order to request a JWT, a user account must have already been created through the web application.

### Endpoints
##### POST /api/jwt
```yml
description:
  - Receive an API key to be used in future API requests
parameters:
  body:
    - username: string
    - password: string
returns:
  - key: string
```
##### GET /api/posts
```yml
description:
  - Get all posts in the system
headers:
  - Authroization: 'Bearer <key>'
returns:
  - [{
       - content:  string
       - author:   string
       - date:     string
       - likes:    int
       - comments: []Comment
       - id:       string
    }]
```
##### GET /api/post/\<id\>
```yml
description:
  - Get a specific post by its id
headers:
  - Authorization: 'Bearer <key>'
parameters:
  url:
    - id: int
returns:
  - {
      - content:  string
      - author:   string
      - date:     string
      - likes:    int
      - comments: []Comment
      - id:       string
    }
```
##### POST /api/post
```yml
description:
  - Create a new blog post
headers:
  - Authorization: 'Bearer <key>'
parameters:
  body:
    - content: string
  returns:
    - message: string
    - post_id: string
```
##### POST /api/comment/<id>
```yml
description:
  - Create a new comment on a blog post
headers:
  - Authorization: 'Bearer <key>'
parameters:
  body:
    - content: string
  url:
    - id: int
returns:
  - message: string
  - comment_id: string
```
##### DELETE /api/post/<id>
```yml
desription:
  - Delete a post with a given id
headers:
  - Authorization: 'Bearer <key>'
parameters:
  url:
    - id: int
returns:
  - message: string
```
##### DELETE /api/comment/\<id\>
```yml
description:
  - Delete a comment with a given id
headers:
  - Authorization: 'Bearer <key>'
parameters:
  url:
    - id: string
  returns:
    - message: string
```
