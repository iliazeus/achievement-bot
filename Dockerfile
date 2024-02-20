FROM golang:1.22-bookworm as golang
WORKDIR /app

RUN apt-get update && apt-get install -y apt-file libwebp-dev

COPY ["go.mod", "go.sum", "./"]
RUN go mod download && go mod verify

COPY ./ ./
RUN CGO_ENABLED=1 go build -o app ./

FROM gcr.io/distroless/base-debian12
COPY --from=golang /usr/lib/*/libwebp*.so* /usr/lib/
COPY --from=golang /app/app /achievement-bot
CMD ["/achievement-bot"]
