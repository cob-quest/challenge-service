FROM golang:alpine as build

WORKDIR /service

COPY . .

RUN go mod download

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o assignment

FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=build /service/assignment .

CMD [ "./assignment" ]
