# Golang REST API with MySQL

MySQL Community Server 5.7.18 + go v1.8.1


## Deployment

```bash
$ docker run --rm -it -v "$GOPATH":/gopath -v "$(pwd)":/app -e "GOPATH=/gopath" -w /app golang:1.4.2 sh -c 'CGO_ENABLED=0 go build -a --installsuffix cgo --ldflags="-s" -o your_binary_name'
```