install:
	go mod tidy
	go install github.com/air-verse/air@latest
	npm install
	mkdir -p .htgo
	touch .htgo/keep

build:
	go run cmd/build/main.go
	HTGO_ENV=production go build -ldflags='-s -w' -o dist/app cmd/app/main.go

start:
	HTGO_ENV=production GIN_MODE=release ./dist/app

dev:
	air \
	--build.cmd "go build -o tmp/bin/main cmd/dev/main.go" \
	--build.bin "tmp/bin/main" \
	--build.delay "100" \
	--build.exclude_dir "app" \
	--build.exclude_dir ".htgo" \
	--build.include_ext "go" \
	--build.stop_on_error "false" \
	--misc.clean_on_exit true \
	--screen.clear_on_rebuild true \
	--log.main_only true
