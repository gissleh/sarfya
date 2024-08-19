FROM docker.io/library/golang:1.23-alpine AS builder
WORKDIR /project
COPY . .
RUN go get
RUN go run ./cmd/sarfya-generate-json/
RUN go build ./cmd/sarfya-prod-server

FROM docker.io/library/alpine:latest
WORKDIR /project
COPY --from=builder /project/sarfya-prod-server /root/.fwew/dictionary-v2.txt /project/data-compiled.json ./
CMD ./sarfya-prod-server
