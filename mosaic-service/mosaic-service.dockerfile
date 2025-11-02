FROM alpine:latest

RUN mkdir /app

COPY ./build/linux/mosaicApp /app

CMD /app/mosaicApp -redis ${REDIS_URL}

# ---- Build stage ----
#FROM ghcr.io/hybridgroup/opencv:4.12.0 AS build
#RUN apt-get update
#RUN go install golang.org/dl/go1.25.1@latest && /go/bin/go1.25.1 download
#RUN git clone https://github.com/hybridgroup/gocv.git
#cd gocv
#make install