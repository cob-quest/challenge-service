FROM golang:alpine as build

WORKDIR /app

COPY . .

RUN go mod download

RUN GOOS=linux go build -ldflags="-w -s" -o challenge

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=build /app/challenge .

CMD [ "./challenge" ]