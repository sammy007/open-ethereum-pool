FROM node:11.1.0-alpine

RUN apk add git

RUN mkdir -p /app
WORKDIR /app

COPY ./ /app

VOLUME /app/www/dist

RUN npm install -g ember-cli@2.9.1
RUN npm install -g bower
RUN npm install
RUN bower --allow-root install

CMD ./build.sh
