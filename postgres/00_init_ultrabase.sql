\set pw `echo $UBERBASE_POSTGRES_PASSWORD`

create role uberbase with password pw NOSUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
create role authelia with password pw NOSUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
create role anon NOSUPERUSER LOGIN;
create database uberbase with owner uberbase;
create database authelia with owner authelia;
grant all on schema uberbase.public to uberbase;
grant all on schema authelia.public to authelia;
grant read on schema uberbase.public to anon;

