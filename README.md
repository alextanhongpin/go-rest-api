# Golang REST API with MySQL

MySQL Community Server 5.7.18 + go v1.8.1


## Deployment
 
```bash
$ docker run --rm -it -v "$GOPATH":/gopath -v "$(pwd)":/app -e "GOPATH=/gopath" -w /app golang:1.4.2 sh -c 'CGO_ENABLED=0 go build -a --installsuffix cgo --ldflags="-s" -o your_binary_name'
```



## Run Docker MySql

```bash
$ docker run --rm --name demo-mysql -it -p 3306:3306 -e MYSQL_ROOT_PASSWORD=123456 -d mysql
```


```bash
$ docker exec -it some-mysql bash
```

## Running a local example without docker

```bash
$ go run main.backup.go
```
