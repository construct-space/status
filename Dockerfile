# Stage 1: Build Vue frontend
FROM node:22-alpine AS frontend
WORKDIR /app/ui
COPY ui/package.json ./
RUN npm install
COPY ui/ .
RUN npm run build

# Stage 2: Build Go binary with embedded frontend
FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY main.go .
COPY --from=frontend /app/web ./web
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o construct-status .

# Stage 3: Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /app/construct-status .
ENV PORT=80
EXPOSE 80
CMD ["./construct-status"]
