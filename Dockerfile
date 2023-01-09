FROM golang:1.19-alpine AS builder
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /app

COPY . .

RUN go mod download
RUN go mod verify

RUN go build -o /godocker

FROM scratch

WORKDIR /

COPY --from=builder /godocker /godocker

ENV NODOTENV=1
ENTRYPOINT ["/godocker"]