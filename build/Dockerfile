FROM       golang:1.16
WORKDIR    /api
COPY       go.mod .
COPY       go.sum .
RUN        go mod download
COPY       . .
RUN        go build -o api ./cmd/main.go
ENTRYPOINT ["./api"]
