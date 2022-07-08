FROM golang

WORKDIR /usr/src/app
EXPOSE 8080

COPY . .
RUN go mod download && go mod verify                      && \
    git clone https://gitlab.sas.com/joboon/adcs-shim.git && \
    cp adcs-shim/cli/linux/amd64/adcscli /usr/bin         && \
    adcscli generate --common-name kindapp                   \
        --san-dns localhost --san-dns 127.0.0.1           && \
    mv server.crt server.key security/                    && \
    rm -rf ca.crt adcscli/                                && \
    go build main.go

CMD ["./main"]
