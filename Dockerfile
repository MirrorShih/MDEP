FROM python:3.10 AS base
RUN git clone https://github.com/radareorg/radare2
RUN radare2/sys/install.sh

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
RUN git clone https://github.com/Y3NH0/GraphTheoryDetector.git
RUN make -C GraphTheoryDetector/
RUN pip install -r GraphTheoryDetector/requirements.txt
ENTRYPOINT ["./main"]
EXPOSE 8000