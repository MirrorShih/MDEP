FROM python:3.10 AS base
RUN git clone https://github.com/radareorg/radare2
RUN radare2/sys/install.sh
RUN git clone https://github.com/Y3NH0/GraphTheoryDetector.git
RUN make -C GraphTheoryDetector/
RUN pip install -r GraphTheoryDetector/requirements.txt

FROM golang:1.20.2-alpine3.17 AS builder
WORKDIR ./src
COPY ./go.mod .
COPY ./go.sum .
RUN go mod download
COPY . /go/src
RUN go build ./main.go

FROM base
COPY --from=builder /go/src ./src
RUN mv ./GraphTheoryDetector/ ./src/GraphTheoryDetector
RUN mkdir ./home/MDEP/ && mkdir ./home/MDEP/upload
WORKDIR ./src
ENTRYPOINT ["./main"]
EXPOSE 8000