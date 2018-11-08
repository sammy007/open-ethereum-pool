FROM golang:1.11

RUN mkdir -p /app
WORKDIR /app

COPY ./ /app

RUN make

CMD ./build/bin/open-ethereum-pool ./config.json
