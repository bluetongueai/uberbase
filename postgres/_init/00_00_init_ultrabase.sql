create role uberbase with password 'uberbase' NOSUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
create role logto with password 'logto' NOSUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
create role anon NOSUPERUSER LOGIN;

create database logto with owner logto;

\c logto
grant all on schema public to postgres;
grant all on schema public to logto;
