FROM alpine:latest

RUN mkdir /app

COPY ./build/linux/dlApp /app

CMD /app/dlApp -nats ${NATS_URL} -redis ${REDIS_URL}