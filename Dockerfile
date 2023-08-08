FROM golang:1.19.12-alpine3.18 as builder
WORKDIR /app
ENV GOOS=linux GOARCH=amd64
COPY go.mod go.sum ./
RUN  go mod download
COPY . .
RUN go vet ./... && go build -o main main.go
RUN apk add curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz

FROM alpine:latest AS release
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/app.env app.env
COPY --from=builder /app/migrate ./migrate
COPY db/migration ./migration
COPY ./start.sh .
COPY ./wait-for.sh .
# https://docs.docker.com/engine/reference/builder/#entrypoint
# ENTRYPOINT ["./start.sh"]
# CMD ["./main"]
ENTRYPOINT ["/app/start.sh"]
CMD ["/app/main"]