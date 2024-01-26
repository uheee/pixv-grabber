# PIXIV Grabber
For unknown reasons, some of your favorites in PIXIV will be masked. This tool is used to backing up your favorite content from PIXIV.



## Installation

### Docker

I recommend deploying it using [docker](https://www.docker.com/). You can configure it with a `docker-compose.yaml`:

```yaml
version: '3'
services:
  grabber:
    image: snowind/pixiv-grabber:v0.1
    volumes:
      # Configuration file
      - ./config.toml:/app/config.toml
      # Storage folder
      - ./output:/app/output
    restart: unless-stopped
    logging:
      driver: "json-file"
      options:
        max-size: "1m"
```

And `config.toml`:

```toml
[log]
level = 'warn'

[job]
# CRON expression for sync
cron = "0 0/4 * * *"
# Your user id
user = '29097741'
# PIXIV API version
version = '3070e6c0783871cc6f72023bc91e05ea646a6005'
# Your cookie, you can find it in your browser
cookie = 'YOUR COOKIE'
# API parameter of 'lang'
lang = 'zh'
# API parameter of 'limit', it controls page size
limit = 100
# Output folder
output = 'output'
```

Put them like this:

```
.
├── docker-compose.yaml
├── config.toml
└── output/
    └── ...
```

At last, run this command:

```shell
docker compose up -d
```



Have a nice journey!
