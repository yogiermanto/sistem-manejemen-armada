FROM golang:1.27.4-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ ./cmd/...

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /out/ /usr/local/bin/
USER nobody
EXPOSE 8080
CMD ["api"]
