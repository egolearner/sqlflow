#!/bin/bash
# Copyright 2019 The SQLFlow Authors. All rights reserved.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

function sleep_until_mysql_is_ready() {
  until mysql -u root -proot --host 127.0.0.1 --port 3306 -e ";" ; do
    sleep 1
    read -p "Can't connect, retrying..."
  done
}


function populate_example_dataset() {
  sleep_until_mysql_is_ready
  # FIXME(typhoonzero): should let docker-entrypoint.sh do this work
  for f in /docker-entrypoint-initdb.d/*; do
    cat $f | mysql -uroot -proot --host 127.0.0.1  --port 3306
  done
}

set -e

service mysql start
sleep 1
populate_example_dataset

go generate ./...
go get -v -t ./...
go install ./...

DATASOURCE="mysql://root:root@tcp(127.0.0.1:3306)/?maxAllowedPacket=0"

sqlflowserver --datasource=${DATASOURCE} &

# e2e test for standar SQL
SQLFLOW_SERVER=localhost:50051 ipython sql/python/test_magic.py
# e2e test for xgboost train and prediciton SQL.
SQLFLOW_SERVER=localhost:50051 ipython sql/python/test_magic_xgboost.py
