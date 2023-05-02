# Stage[0]: Build goose binaries
FROM golang:1.18.0 AS build

RUN go install github.com/pressly/goose/v3/cmd/goose@latest
COPY ./goose /docker-entrypoint-initdb.d/goose
RUN (cd /docker-entrypoint-initdb.d/goose/migrations/catalog && goose fix) 
RUN (cd /docker-entrypoint-initdb.d/goose/migrations/blob && goose fix) 
RUN (cd /docker-entrypoint-initdb.d/goose/; go build -mod vendor -o /usr/bin/goose.catalog -tags catalog)
RUN (cd /docker-entrypoint-initdb.d/goose/; go build -mod vendor -o /usr/bin/goose.blob -tags blob)
RUN chmod +x /usr/bin/goose*

# Stage[1]: Apply migration
FROM postgres:12

COPY --from=build /usr/bin/goose.catalog /usr/bin/goose.catalog
COPY --from=build /usr/bin/goose.blob /usr/bin/goose.blob
# RUN chmod +x /usr/bin/goose

COPY --from=build /docker-entrypoint-initdb.d/goose /docker-entrypoint-initdb.d/goose

COPY ./docker-entrypoint-initdb.d /docker-entrypoint-initdb.d
# RUN chmod -R +r ./docker-entrypoint-initdb.d; chmod +x ./docker-entrypoint-initdb.d/catalog~20210104.pgdump.d
RUN (apt-get update && apt-get install -y postgresql-plperl-12 postgresql-contrib-12 golang)
# RUN su - postgres -c postgres
# RUN until [ -f /tmp/init.lock ] do sleep 10; echo "WAIT for lockfile" done

# # Stage[2]: Copy migrated database to clean final container
# FROM postgres:12

# RUN apt-get update -y && apt-get install -y postgresql-plperl-12 postgresql-contrib-12