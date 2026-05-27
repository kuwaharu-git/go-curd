FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /todo-app ./main.go

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /todo-app /todo-app
COPY templates ./templates
COPY static ./static
EXPOSE 8080
CMD ["/todo-app"]
