FROM golang:1.24-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o main

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/main .
EXPOSE 8081
CMD ["./main"]