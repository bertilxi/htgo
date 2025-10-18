install:
	go mod tidy
	cd app && npm install
	mkdir -p .htgo
	touch .htgo/keep

build:
	go run cmd/build/main.go
	HTGO_ENV=production go build -ldflags='-s -w' -o dist/app cmd/app/main.go

start:
	HTGO_ENV=production GIN_MODE=release ./dist/app

dev:
	go run cmd/dev/main.go
