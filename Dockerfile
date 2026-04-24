FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /scheduling-service ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /scheduling-service /app/scheduling-service
USER 65532:65532
EXPOSE 8081
ENTRYPOINT ["/app/scheduling-service"]
