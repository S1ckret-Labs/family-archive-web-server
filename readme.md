# Family archive web server

## About

Web server that handles business logic such as:

- saving files metadata into a database
- manipulating file tree
- scheduling archival jobs

## Setup on Linux

1. Install `go`
2. Create your own `.env` file using `.template.env` example
3. Enable auto export

```shell
set -o allexport
```

4. Source env variables

```shell
. .env
```

5. Build the server

```shell
go install github.com/swaggo/swag/cmd/swag@latest
swag init --parseDependency --parseInternal
# Remove LeftDelim and RightDelim from docs/docs.go
go build
```


6. Run the server

```shell
./family-archive-web-server
```