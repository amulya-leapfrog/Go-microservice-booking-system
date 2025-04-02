# build a small docker image
FROM alpine:latest

RUN mkdir /app

COPY loggerApp /app

CMD ["/app/loggerApp"]