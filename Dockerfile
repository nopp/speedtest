FROM golang:1.24 AS build

WORKDIR /app

COPY . .

RUN go build -o speedtest .

FROM debian:bookworm-slim

COPY --from=build /app/speedtest /speedtest

EXPOSE 9108

CMD ["/speedtest"]
