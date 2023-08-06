FROM golang:1.19.12-alpine3.18 as builder
WORKDIR /app
ENV GOOS=linux GOARCH=amd64
COPY go.mod go.sum ./
RUN  go mod download
COPY . .
RUN go vet ./... && go build -o main main.go
RUN ls && pwd

FROM alpine:latest AS release
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/app.env app.env
EXPOSE 8080
ENTRYPOINT ["./main"]