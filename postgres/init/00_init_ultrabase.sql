\set pw `echo \'$UBERBASE_POSTGRES_PASSWORD\'`

create role uberbase with password :pw NOSUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
create role zitadel with password 'zitadel' NOSUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
create role anon NOSUPERUSER LOGIN;

-- create database uberbase with owner uberbase;
create database zitadel with owner zitadel;

\c zitadel
grant all on schema public to zitadel;

\c postgres
