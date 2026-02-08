# Use the official Golang image
FROM golang:1.24-alpine

# Install build essentials and git
# Adding --update helps refresh the package index
RUN apk add --no-cache --update git build-base

# 1. Set the environment to be explicit
ENV GO111MODULE=on
ENV GOPROXY=https://proxy.golang.org,direct

# Install Air
RUN go install github.com/air-verse/air@v1.61.7

WORKDIR /app

# Pre-copy modules to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Air will handle the building and running
CMD ["air", "-c", ".air.toml"]