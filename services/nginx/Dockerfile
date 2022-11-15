FROM node:16 AS BUILD

COPY ./frontend /tmp/build/
RUN (cd /tmp/build; yarn install; yarn build)

FROM ubuntu:jammy AS WGET
# Install Dockerize
RUN apt-get update && apt-get install -y wget unzip
ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-linux-amd64-$DOCKERIZE_VERSION.tar.gz

FROM nginx:mainline-alpine

# Copy Dockerize
COPY --from=WGET /usr/local/bin/dockerize /usr/local/bin/dockerize
# Copy frontdend
COPY --from=BUILD /tmp/build/dist/ /usr/share/nginx/html/

COPY ./conf.d/ /etc/nginx/conf.d/

ENTRYPOINT ["dockerize", "-template", "/etc/nginx/conf.d/default.conf.tmpl:/etc/nginx/conf.d/default.conf"]
CMD ["nginx", "-g", "daemon off;"]