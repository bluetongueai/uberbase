create role anon NOSUPERUSER NOLOGIN;
create role admin NOSUPERUSER NOLOGIN;
create role authenticator noinherit login password 'postgrest-secret-user-password';

grant anon to authenticator;
grant admin to authenticator;

create database uberbase with owner postgres encoding 'UTF8';
create database uberbase_fusionauth with owner postgres encoding 'UTF8';
