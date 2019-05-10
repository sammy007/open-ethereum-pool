# Build Geth in a stock Go builder container
FROM golang:1.10-alpine as construction

RUN apk add --no-cache make gcc musl-dev linux-headers

ADD . /truechain-engineering-code
RUN cd /truechain-engineering-code && make getrue

# Pull Geth into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=construction /truechain-engineering-code/build/bin/getrue /usr/local/bin/
CMD ["getrue"]

EXPOSE 8545 8545 9215 9215 30310 30310 30311 30311 30313 30313
ENTRYPOINT ["getrue"]


