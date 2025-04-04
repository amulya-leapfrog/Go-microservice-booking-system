FROM alpine:latest

RUN mkdir /app

COPY reservationApp /app

CMD ["/app/reservationApp"]