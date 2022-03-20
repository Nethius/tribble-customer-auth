FROM golang:1.16

WORKDIR /trimble-auth

COPY . .

RUN make build
