FROM golang:1.19-alpine AS builder

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