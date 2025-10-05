FROM ubuntu:latest

RUN mkdir /app

COPY ./build/linux/brokerApp /app

CMD /app/brokerApp -p ${PORT}