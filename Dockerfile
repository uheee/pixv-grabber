FROM golang:1.21-alpine as build
ENV GO111MODULE=on

WORKDIR /app
COPY . .
RUN go build -ldflags '-s -w' -o 'target/pgr'

FROM alpine as final
WORKDIR /app
COPY --from=build /app/target .
ENTRYPOINT ["./pgr"]