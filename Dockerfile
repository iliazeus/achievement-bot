FROM golang:1.22-alpine as build
WORKDIR /app

RUN apk add gcc musl-dev libwebp-dev libwebp-static

COPY ["go.mod", "go.sum", "./"]
RUN go mod download && go mod verify

COPY ./ ./
RUN CGO_ENABLED=1 CGO_LDFLAGS='-lsharpyuv' go build \
  --ldflags '-linkmode external -extldflags "-static"' \
  -o app ./

FROM scratch
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /app/app /app/achievement-bot
CMD ["/app/achievement-bot"]
