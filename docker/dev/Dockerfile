FROM golang:1.17.4 as builder

ARG APP_VERSION

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal/ ./internal

RUN GOARCH=amd64 GOOS=linux go build -ldflags "-X main.version=${APP_VERSION}" -o server ./cmd/server

FROM scratch

COPY --from=builder /server /server

EXPOSE 3009

CMD ["/server"]

