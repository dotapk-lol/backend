FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /dueld ./cmd/dueld
FROM scratch
COPY --from=build /dueld /dueld
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/dueld"]
