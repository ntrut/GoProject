FROM golang:1.13.1

WORKDIR /go/src/app

COPY go.mod ./
COPY go.sum ./
COPY *.go ./

RUN go build -o /GoProject

EXPOSE 8080

CMD ["/GoProject"]