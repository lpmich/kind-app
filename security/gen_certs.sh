#! /bin/bash

# Generate certificate authority data
openssl req -new -sha256 -nodes -newkey rsa:4096 -keyout CA.key -out CA.csr \
    -subj /C=US/ST=none/L=none/O=none/OU=none/CN=none
openssl x509 -req -sha256 -extfile ./security/x509.ext -extensions ca -in CA.csr -signkey CA.key \
    -days 1095 -out CA.pem

# Generate server key and create CSR
openssl req -new -sha256 -nodes -newkey rsa:4096 -keyout server.key -out server.csr \
    -subj /C=US/ST=NA/L=NA/O=NA/OU=NA/CN=server

# Self-sign CSR
openssl x509 -req -sha256 -CA CA.pem -CAkey CA.key -days 1095 -CAcreateserial -CAserial CA.srl \
   -extfile ./security/x509.ext -extensions server -in server.csr -out server.pem

# Clean up directory
mv server.key server.pem ./security
rm ./server.* ./CA.*
