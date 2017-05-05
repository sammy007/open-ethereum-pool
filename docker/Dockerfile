FROM ubuntu:16.04
RUN apt-get update && apt-get -y install curl && curl -sL https://deb.nodesource.com/setup_6.x | bash -
RUN apt-get update && apt-get -y install build-essential golang nodejs git

RUN git config --global http.https://gopkg.in.followRedirects true

WORKDIR /opt

RUN git clone https://github.com/Musicoin/go-musicoin.git

WORKDIR /opt/go-musicoin

RUN make gmc

RUN chmod a+x /opt/go-musicoin/build/bin/gmc
RUN ln -s /opt/go-musicoin/build/bin/gmc /usr/bin/geth
RUN ln -s /opt/go-musicoin/build/bin/gmc /usr/bin/gmc

WORKDIR /opt

RUN git config --global http.https://gopkg.in.followRedirects true
RUN git clone --branch release/music --single-branch https://github.com/bulktrade/open-ethereum-pool.git

WORKDIR /opt/open-ethereum-pool

RUN make

WORKDIR /opt/open-ethereum-pool/www

RUN npm install -g ember-cli@2.9.1 bower && npm install && bower --allow-root install && ./build.sh

WORKDIR /

RUN apt-get update && apt-get -y install nginx
RUN apt-get update && apt-get -y install supervisor

ENV DEBIAN_FRONTEND noninteractive

COPY nginx.conf /etc/nginx/sites-enabled/default
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf
COPY pool.conf /etc/security/limits.d/pool.conf

CMD ["/usr/bin/supervisord"]

VOLUME ["/root/.musicoin"]

EXPOSE 8008 80
