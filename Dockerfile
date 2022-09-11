FROM golang

WORKDIR /usr/src/app
EXPOSE 8080

COPY . .
RUN go mod download && \
    go mod verify   && \
    go build main.go

CMD ["./main"]
