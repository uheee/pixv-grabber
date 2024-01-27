build:
	go build -ldflags '-s -w' -o 'target/grabber' github.com/uheee/pixiv-grabber/cmd/sync