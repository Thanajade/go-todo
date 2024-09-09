go get github.com/gorilla/mux
go get github.com/golang-jwt/jwt/v5

go test main.go main_test.go -v

go install github.com/codegangsta/gin@latest

gin run main.go --appPort 8000
