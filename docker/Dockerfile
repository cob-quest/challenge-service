FROM golang:1.20

WORKDIR /app

ENV CGO_ENABLED 1

COPY ./go.mod ./

COPY ./go.sum ./

RUN go mod download

COPY . ./

RUN go build -o /cob/assignment-service

# EXPOSE 8080

CMD [ "/cob/assignment-service" ]
