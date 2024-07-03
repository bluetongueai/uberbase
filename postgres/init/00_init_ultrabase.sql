\set pw `echo \'$UBERBASE_POSTGRES_PASSWORD\'`

create role uberbase with password :pw NOSUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
create role casdoor with password :pw NOSUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
create role anon NOSUPERUSER LOGIN;

-- create database uberbase with owner uberbase;
create database casdoor with owner casdoor;

\c casdoor
grant all on schema public to casdoor;

