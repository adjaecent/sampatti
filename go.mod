module github.com/adjaecent/sampatti

go 1.25.0

require (
	github.com/adjaecent/unofficial-kuvera-api v0.0.0-20241008000000-000000000000
	github.com/adjaecent/unofficial-stockal-api v0.0.0-20241008000000-000000000000
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/sessions v1.3.0
	github.com/joho/godotenv v1.5.1
	github.com/mattn/go-sqlite3 v1.14.19
	golang.org/x/oauth2 v0.18.0
)

require (
	cloud.google.com/go/compute v1.20.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

replace github.com/adjaecent/unofficial-kuvera-api => ../unofficial-kuvera-api

replace github.com/adjaecent/unofficial-stockal-api => ../unofficial-stockal-api
