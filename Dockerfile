FROM golang:latest AS build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cmd/gophkeeper/bin/main ./cmd/gophkeeper/

FROM alpine:latest
WORKDIR /gophkeeper
RUN mkdir /gophkeeper/logs
COPY --from=build /build/migrations /gophkeeper/migrations
COPY --from=build /build/cmd/gophkeeper/bin/main .
CMD ["/gophkeeper/main"]