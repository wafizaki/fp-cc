FROM golang:1.24-alpine AS build
WORKDIR /app

# Install build dependencies for cgo and sqlite3
RUN apk add --no-cache gcc musl-dev sqlite-dev

ENV CGO_ENABLED=1
COPY . .
RUN go build -o main

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache sqlite-libs
COPY --from=build /app/main .
COPY --from=build /app/frontend ./frontend
EXPOSE 8081
CMD ["./main"]