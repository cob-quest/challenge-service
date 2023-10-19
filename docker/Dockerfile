FROM golang:alpine as build

WORKDIR /service

COPY . .

RUN go mod download

RUN GOOS=linux go build -ldflags="-w -s" -o assignment

FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=build /service/assignment .

CMD [ "./assignment" ]
