#FROM golang:1.12-stretch as build
#
#WORKDIR /tp_db_forum
#COPY . .
##RUN CGO_ENABLED=0
#RUN go get -d && go build -v

FROM ubuntu:18.04

# install PostgreSQL
ENV PGSQLVER 10
RUN apt-get update &&\
    apt-get install -y postgresql-$PGSQLVER postgresql-contrib &&\
    apt-get install -y git &&\
    apt-get install -y wget

# create PostgreSQL database
USER postgres
RUN    /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    # psql --command "CREATE DATABASE docker WITH ENCODING 'UTF8';" &&\
    createdb -O docker docker &&\
    psql -d docker -c "CREATE EXTENSION IF NOT EXISTS citext;" &&\
    # psql --command "alter system set max_connections = 1000;" &&\
    # psql --command "alter system set lc_collate = 'en_US.UTF-8';"  &&\
    /etc/init.d/postgresql stop
RUN echo "host all  all    0.0.0.0/0  md5" >> /etc/postgresql/$PGSQLVER/main/pg_hba.conf &&\
    echo "listen_addresses='*'" >> /etc/postgresql/$PGSQLVER/main/postgresql.conf &&\
    echo "default_text_search_config = 'pg_catalog.english'" >> /etc/postgresql/$PGSQLVER/main/postgresql.conf

# RUN initdb --locale=en_US

# Expose the PostgreSQL port
EXPOSE 5432

# open Postgres for network

# Golang installing

ENV GOVERSION 1.12
USER root
RUN wget https://storage.googleapis.com/golang/go$GOVERSION.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go$GOVERSION.linux-amd64.tar.gz
ENV GOROOT /usr/local/go
ENV GOPATH /opt/go
ENV PATH $GOROOT/bin:$GOPATH/bin:$PATH
#RUN go build /tp_db_forum .
EXPOSE 5000

WORKDIR /tp_db_forum
COPY . .

#RUN echo "./config/postgresql.conf" >> /etc/postgresql/$PGSQLVER/main/postgresql.conf
#COPY --from=build /tp_db_forum .

CMD service postgresql start && go run main.go
