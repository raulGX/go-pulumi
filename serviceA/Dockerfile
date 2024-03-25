
FROM golang:1.22.1-bookworm as builder

WORKDIR /app
COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -v -o server

FROM debian:bookworm-slim
COPY --from=builder /app/server /app/server

CMD ["/app/server"]

# [END run_helloworld_dockerfile]
# [END cloudrun_helloworld_dockerfile]