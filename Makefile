build:
	go build -ldflags '-s -w' -o 'target/grabber' github.com/uheee/pixiv-grabber/cmd/sync

migrators:
	go build -ldflags '-s -w' -o 'target/l2m' github.com/uheee/pixiv-grabber/cmd/migrators/l2m