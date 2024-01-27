FROM golang:1.21-alpine as build
RUN apk add --no-cache --update build-base
ENV GO111MODULE=on
WORKDIR /app
COPY . .
RUN make

FROM alpine as final
WORKDIR /app
COPY --from=build /app/target .
ENTRYPOINT ["./grabber"]