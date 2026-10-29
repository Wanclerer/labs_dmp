FROM golang:1.27-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/shortener ./cmd/shortener

FROM alpine:3.21

RUN adduser -D -u 10001 app
USER app
COPY --from=build /out/shortener /usr/local/bin/shortener

EXPOSE 8080

# Именно exec-форма. В shell-форме бинарник запускается через оболочку,
# сигнал до него не доходит, и корректная остановка молча не работает.
CMD ["/usr/local/bin/shortener"]
