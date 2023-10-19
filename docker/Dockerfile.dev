FROM golang:alpine

WORKDIR /app

RUN go install github.com/cosmtrek/air@latest

COPY go.* .

RUN go mod download

COPY . .

EXPOSE 5050

CMD [ "air", "-d"]