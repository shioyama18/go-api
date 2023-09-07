FROM golang:1.21
WORKDIR /go/src/go-api
ENV GOPROXY direct
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/server/main.go ./
COPY handlers ./handlers
COPY models ./models
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .
FROM alpine:latest
WORKDIR /root
COPY --from=0 /go/src/go-api/app .
CMD ["./app"]