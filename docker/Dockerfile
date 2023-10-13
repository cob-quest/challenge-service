FROM golang:alpine as build

WORKDIR /service

COPY . .

RUN go mod download

RUN go build -o assignment

FROM alpine:latest

WORKDIR /

COPY --from=build /service/assignment .

CMD [ "./assignment" ]
