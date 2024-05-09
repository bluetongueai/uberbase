\set pw `echo \'$UBERBASE_POSTGRES_PASSWORD\'`

create role uberbase with password :pw NOSUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
create role authelia with password :pw NOSUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
create role anon NOSUPERUSER LOGIN;
-- create database uberbase with owner uberbase;
create database authelia with owner authelia;
\c uberbase
grant all on schema public to uberbase;
grant usage on schema public to anon;
\c authelia
grant all on schema public to authelia;

