#!/bin/bash
mkdir -p db_init
cp schema.sql db_init/init.sql
docker run -d --rm -p 5433:5432 --name pdb -e POSTGRES_PASSWORD=${POSTGRES_PASSWORD} -v $(pwd)/db_init:/docker-entrypoint-initdb.d postgres