FROM golang:1.16-alpine3.13 as builder
WORKDIR /app
ADD . .
RUN go build -o main main.go

FROM alpine:3.13
WORKDIR /app
COPY --from=builder /app/main .
COPY alarm.csv .
ADD config config
EXPOSE 55555
CMD ["./main"]

