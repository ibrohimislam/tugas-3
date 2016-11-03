# TermChat

TermChat ini menggunakan server RabbitMQ `bunny.ibrohim.me`.

## Requirement

- [Go](https://golang.org/doc/install) is Installed
- Can connect to `ibrohim.me:5672`

## Usage

`go get github.com/streadway/amqp`\
`go get github.com/ibrohimislam/tugas-3`\
`cd $GOPATH/src/github.com/ibrohimislam/tugas-3`
- untuk menjalankan server `go run server/main/main.go`
- untuk menjalankan client `go run client/main.go`

## Commands

Berikut adalah perintah yang dapat digunakan sebelum login:

```
register [username] [password]
login [username] [password]
```

Perintah `register` digunakan untuk melakukan registrasi ke server. Perintah `login` digunakan untuk melakukan login ke server dan mendapatkan token. CLI Client secara otomatis menmanage token untuk perintah-perintah berikutnya.

Berikut adalah perintah yang dapat digunakan setelah login:
```
addfriend [username]
creategroup [groupname]
addmember [groupname] [username]
addfriend [username]
removemember [groupname] [username]
leave [groupname]
sendtogroup [groupname] [message]
sendtouser [username] [message]
```
