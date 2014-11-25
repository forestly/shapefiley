#!/usr/bin/env bash

cd /tmp/shapefiley/
unzip -a $1
/usr/local/bin/shp2pgsql -s 4326 -I -c -W UTF-8 $2 $2 > $2.sql
psql shapefiley_work_development < $2.sql
