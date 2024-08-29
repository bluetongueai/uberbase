create role anon NOSUPERUSER NOLOGIN;
create role admin NOSUPERUSER NOLOGIN;
create role authenticated NOSUPERUSER NOLOGIN;
create role authenticator noinherit login password 'postgres-secret-user-password';

grant anon to authenticator;
grant admin to authenticator;
grant authenticated to authenticator;

grant all on schema public to admin;

create database fusionauth with owner postgres encoding 'UTF8';
