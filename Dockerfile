FROM golang:alpine AS binarybuilder
WORKDIR /poorsquad
COPY . .
RUN apk --no-cache --no-progress add --virtual build-deps build-base git linux-pam-dev \
    && go mod tidy -v \
    && go build -o dashboard -ldflags="-s -w" cmd/dashboard/main.go
FROM alpine:latest
RUN apk --no-cache --no-progress add \
    ca-certificates \
    tzdata
WORKDIR /poorsquad
COPY resource /poorsquad/resource
COPY --from=binarybuilder /poorsquad/dashboard ./dashboard

VOLUME ["/poorsquad/data"]
EXPOSE 8080
CMD ["/poorsquad/dashboard"]