FROM golang:alpine

WORKDIR /go/src/app
COPY . .

RUN apk add --no-cache git make build-base
RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 8081

CMD ["gitlab-go-webhook-dashboard"]
