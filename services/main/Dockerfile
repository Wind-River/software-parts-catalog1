FROM golang:1.18 as BUILD
# Install libarchive and build binary

##
# Install libarchive
RUN wget https://www.libarchive.org/downloads/libarchive-3.5.0.tar.gz
RUN echo "MD5 (libarchive-3.5.0.tar.gz) = 39eea5d132b48af1be4b0f477bad06a8" | md5sum --check -
RUN apt-get update && apt-get install -y build-essential zlib1g-dev libcrypto++-dev libssl-dev liblzma-dev
RUN tar -xvf libarchive-3.5.0.tar.gz
RUN cd libarchive-3.5.0 && ./configure --prefix=/install/ && make install
RUN cp /install/include/* /usr/local/include/ -r
RUN cp /install/lib/* /usr/local/lib/ -r
RUN cp /install/share/* /usr/local/share/ -r
##

##
# Build Code
RUN mkdir -p /tmp/build

COPY ./ /tmp/build/

RUN cd /tmp/build/; GOPRIVATE="gitlab.devstar.cloud" go build -o server-linux-amd64.bin
##

FROM ubuntu:jammy as WGET
# Install dockerize and promtail

RUN apt-get update && apt-get install -y wget unzip
RUN apt-get install -y bzip2
ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz
RUN wget https://github.com/grafana/loki/releases/download/v2.4.1/promtail-linux-amd64.zip && unzip promtail-linux-amd64.zip


FROM ubuntu:jammy as FINAL

# COPY libarchive
COPY --from=BUILD /install/include/ /usr/local/include/
COPY --from=BUILD /install/lib/ /usr/local/lib/
COPY --from=BUILD /install/share/ /usr/local/share/
# COPY libcrypto
COPY --from=BUILD /usr/lib/x86_64-linux-gnu/libcrypto* /usr/lib/x86_64-linux-gnu/
# COPY bzip2 for libarchive
COPY --from=WGET /usr/bin/bzip2 /usr/bin/bzip2
# Copy binary
COPY --from=BUILD /tmp/build/server-linux-amd64.bin /opt/tk/server-linux-amd64.bin
# Copy Dockerize
COPY --from=WGET /usr/local/bin/dockerize /usr/local/bin/dockerize
# Copy promtail
COPY --from=WGET promtail-linux-amd64 /usr/local/bin/promtail


RUN apt-get update -y && apt-get install -y supervisor
RUN apt-get install -y vim

RUN mkdir -p /opt/tk /opt/tk/uploads /var/log/highlander
COPY ./docker/conf.d /etc/supervisor/conf.d
RUN mkdir -p /etc/promtail && mv /etc/supervisor/conf.d/promtail-local-config.yaml /etc/promtail/

COPY ./docker/credentials /home/root/.aws/credentials

# Make directories that may be mounted over
RUN mkdir /opt/tk/blob

COPY ./docker/entrypoint.sh /usr/bin/entrypoint
RUN  chmod +x /usr/bin/entrypoint
COPY ./docker/run.sh /run.sh
RUN chmod +x /run.sh
RUN touch /var/log/run.log /var/log/run.error.log
ENTRYPOINT [ "entrypoint" ]
CMD [ "/run.sh" ]