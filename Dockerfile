FROM golang:1.20

WORKDIR /go-storage

COPY . /go-storage

RUN go mod download
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest