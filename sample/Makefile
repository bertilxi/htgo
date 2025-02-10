install:
	go mod tidy
	go install github.com/air-verse/air@latest
	cd app && npm install

build:
	GIN_MODE=release go run cmd/build/main.go
	GIN_MODE=release go build -o dist/app cmd/app/main.go

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
