create role socratic with password 'socratic' NOSUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
grant all on schema public to socratic;
create role supertokens with password 'supertokens' NOSUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
grant all on schema public to supertokens;
create role coder with password 'coder' NOSUPERUSER CREATEDB CREATEROLE INHERIT LOGIN;
grant all on schema public to coder;
create database socratic_development with owner socratic;
create database socratic_test with owner socratic;
create database socratic_production with owner socratic;
create database supertokens with owner supertokens;
create database coder with owner coder;

