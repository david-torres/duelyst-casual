FROM golang:1.7

RUN mkdir -p /go/src/github.com/david-torres/duelyst-casual
WORKDIR /go/src/github.com/david-torres/duelyst-casual

COPY . /go/src/github.com/david-torres/duelyst-casual

RUN go-wrapper download
RUN go-wrapper install

ENV PORT 3000
EXPOSE 3000

CMD ["go-wrapper", "run"]