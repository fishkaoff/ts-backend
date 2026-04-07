FROM golang:latest AS builder  

WORKDIR /app

COPY go.mod ./

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go mod download

COPY . .

RUN go build  -o ./ts-backend ./cmd/api/main.go

FROM debian:bookworm-slim
COPY --from=builder /app/ts-backend /ts-backend

RUN chmod +x /ts-backend

CMD ["/ts-backend"]