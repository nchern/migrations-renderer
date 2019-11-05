#!/bin/sh

SRC_PATH=${1:-"."}

DB_IMAGE="postgres:11-alpine"
DB_CONTAINER="pg_host"
DB_USER="postgres"
DB_PASSWD="123"

FLYWAY_IMAGE="flyway/flyway:6.0.7-alpine"
FLYWAY_URL="jdbc:postgresql://$DB_CONTAINER/postgres"
FLYWAY_PATH="$(realpath $SRC_PATH)"

[ -d "$FLYWAY_PATH" ] || exit 1

docker run --rm --name $DB_CONTAINER -e "POSTGRES_PASSWORD=$DB_PASSWD" -d $DB_IMAGE 1>&2

sleep 1

docker run --rm --link $DB_CONTAINER -v "$FLYWAY_PATH:/flyway/sql" -it $FLYWAY_IMAGE -url=$FLYWAY_URL -user=$DB_USER -password=$DB_PASSWD migrate 1>&2

docker run --rm --link $DB_CONTAINER -e "PGPASSWORD=$DB_PASSWD" -it $DB_IMAGE pg_dump -s -h $DB_CONTAINER -U $DB_USER

docker stop $DB_CONTAINER 1>&2 
