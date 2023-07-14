FROM python:3.10 AS base
RUN pip install -U r2env \
&& r2env add radare2@git

FROM golang:1.20.2-alpine3.17 AS builder
WORKDIR ./src
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download
COPY . /go/src
RUN go build ./main.go

FROM base
COPY --from=builder /go/src ./src
WORKDIR ./src
ENTRYPOINT ["./main"]
EXPOSE 8000