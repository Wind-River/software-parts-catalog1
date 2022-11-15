#!/usr/bin/env bash

# Copyright (c) 2020 Wind River Systems, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at:
#       http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software  distributed
# under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
# OR CONDITIONS OF ANY KIND, either express or implied.

set -e

dockerize \
    -template /etc/supervisor/conf.d/tk.conf.tmpl:/etc/supervisor/conf.d/tk.conf \
    -template /etc/supervisor/conf.d/tk.toml.tmpl:/etc/supervisor/conf.d/tk.toml \
    -wait tcp://${DB_HOST}:${DB_PORT} \
    -wait tcp://${S3_ENDPOINT} \
    "$@"