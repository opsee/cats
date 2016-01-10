SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET default_tablespace = '';
SET default_with_oids = false;

CREATE EXTENSION "uuid-ossp";

CREATE TYPE relationship_type AS enum (
  'equal',
  'notEqual',
  'empty',
  'notEmpty',
  'contain',
  'notContain',
  'regExp'
);

CREATE TABLE assertions (
  check_id character varying(255) NOT NULL,
  customer_id UUID NOT NULL,
  key character varying(255) NOT NULL,
  relationship relationship_type NOT NULL,
  value character varying(255) NOT NULL,
  operand character varying(255) NOT NULL
);

CREATE INDEX idx_assertions_check_id_and_customer_id ON assertions (check_id, customer_id);
