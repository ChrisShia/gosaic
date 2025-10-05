FROM alpine:latest

RUN mkdir /app

COPY ./build/linux/mosaicApp /app

CMD /app/mosaicApp -redis ${REDIS_URL}