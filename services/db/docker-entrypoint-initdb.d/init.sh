#!/usr/bin/env bash

# Copyright (c) 2020 Wind River Systems, Inc.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at:
#       http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software  distributed
# under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
# OR CONDITIONS OF ANY KIND, either express or implied.

set -e

initdir="/docker-entrypoint-initdb.d"

echo "Connecting to ${POSTGRES_DB} as ${POSTGRES_USER}"


psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE USER hladmin WITH LOGIN;
    CREATE DATABASE catalog;
    CREATE DATABASE blob;
EOSQL


# Run any restorations on catalog
if [ -d $initdir/restore/catalog ]; then
    echo "checking restores for catalog"
    for i in $initdir/restore/cotolag/*; do
        if [[ "$i" == *.sh ]]; then
            echo "running script $i"
            source $i
        else
            echo "restoring to catalog from $i"
            pg_restore -v -e --username "$POSTGRES_USER" --dbname catalog $i
        fi
    done
    unset i
fi

# Run any restorations on blob
if [ -d $initdir/restore/blob ]; then
    echo "checking restores for blob"
    for i in $initdir/restore/blob/*; do
        echo "restoring to blob from $i"
        pg_restore -v -e --username "$POSTGRES_USER" --dbname blob $i
    done
    unset i
fi

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    GRANT ALL PRIVILEGES ON DATABASE catalog TO hladmin;
    ALTER DATABASE catalog OWNER TO hladmin;
    GRANT ALL PRIVILEGES ON DATABASE blob TO hladmin;
    ALTER DATABASE blob OWNER TO hladmin;
EOSQL

# # During init, server only accepts socket connections; flyway being a java application, does not support unix sockets without a socketFactory class
# ## Restart postgres on localhost 
# pg_ctl -D "$PGDATA" -m fast -w stop
# pg_ctl -D "$PGDATA" -o "-c listen_addresses='localhost'" -w start
# while ! pg_isready -h localhost -p 5432 -d highlander
# do 
#     sleep 5
# done
##
echo "Migrating catalog"
goose.catalog --dir $initdir/goose/migrations/catalog --host "/var/run/postgresql" --dbname catalog up
echo "Migrating blob"
goose.blob --dir $initdir/goose/migrations/blob --host "/var/run/postgresql" --dbname blob up
date --utc > /tmp/init.lock