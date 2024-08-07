FROM golang:1.22-alpine AS builder
WORKDIR /project
COPY . .
RUN go get
RUN go run ./cmd/sarfya-generate-json/
RUN go build ./cmd/sarfya-prod-server

FROM alpine:latest
WORKDIR /project
COPY --from=builder /project/sarfya-prod-server /project/dictionary-v2.txt /project/data-compiled.json ./
CMD ./sarfya-prod-server