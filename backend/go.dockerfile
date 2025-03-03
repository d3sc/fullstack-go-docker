FROM golang:1.24-alpine3.20 AS builder

WORKDIR /app

COPY . .

# Download and install depedencies
RUN go get -d -v ./...

# build the go app
# the name in go.mod module is an api
RUN go build -o api .

EXPOSE 8000

CMD ["./api"]