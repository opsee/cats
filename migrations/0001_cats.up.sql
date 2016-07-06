--
-- PostgreSQL database dump
--

-- Dumped from database version 9.5.2
-- Dumped by pg_dump version 9.5.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET search_path = public, pg_catalog;

--
-- Name: relationship_type; Type: TYPE; Schema: public; Owner: bartnet
--

CREATE TYPE relationship_type AS ENUM (
    'equal',
    'notEqual',
    'empty',
    'notEmpty',
    'contain',
    'notContain',
    'regExp',
    'greaterThan',
    'lessThan'
);

--
-- Name: update_time(); Type: FUNCTION; Schema: public; Owner: bartnet
--

CREATE FUNCTION update_time() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
      BEGIN
      NEW.updated_at := CURRENT_TIMESTAMP;
      RETURN NEW;
      END;
      $$;


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: assertions; Type: TABLE; Schema: public; Owner: bartnet
--

CREATE TABLE assertions (
    check_id character varying(255) DEFAULT ''::character varying NOT NULL,
    customer_id uuid NOT NULL,
    key character varying(255) DEFAULT ''::character varying NOT NULL,
    relationship relationship_type NOT NULL,
    value character varying(255) DEFAULT ''::character varying,
    operand character varying(255) DEFAULT ''::character varying
);


--
-- Name: check_state_memos; Type: TABLE; Schema: public; Owner: bartnet
--

CREATE TABLE check_state_memos (
    check_id character varying(255) NOT NULL,
    customer_id uuid NOT NULL,
    bastion_id uuid NOT NULL,
    failing_count integer NOT NULL,
    response_count integer NOT NULL,
    last_updated timestamp with time zone NOT NULL
);


--
-- Name: check_states; Type: TABLE; Schema: public; Owner: bartnet
--

CREATE TABLE check_states (
    check_id character varying(255) NOT NULL,
    customer_id uuid NOT NULL,
    state_id integer NOT NULL,
    state_name character varying(255) NOT NULL,
    time_entered timestamp with time zone NOT NULL,
    last_updated timestamp with time zone NOT NULL,
    failing_count integer NOT NULL,
    response_count integer NOT NULL
);


--
-- Name: checks; Type: TABLE; Schema: public; Owner: bartnet
--

CREATE TABLE checks (
    id character varying(255) NOT NULL,
    "interval" integer,
    target_id character varying(255) NOT NULL,
    check_spec jsonb,
    customer_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    target_name character varying(255),
    target_type character varying(255) NOT NULL,
    execution_group_id uuid NOT NULL,
    min_failing_count integer DEFAULT 1 NOT NULL,
    min_failing_time integer DEFAULT 90 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: credentials; Type: TABLE; Schema: public; Owner: bartnet
--

CREATE TABLE credentials (
    id integer NOT NULL,
    provider character varying(20),
    access_key_id character varying(60),
    secret_key character varying(60),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    customer_id uuid NOT NULL
);


--
-- Name: credentials_id_seq; Type: SEQUENCE; Schema: public; Owner: bartnet
--

CREATE SEQUENCE credentials_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: credentials_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: bartnet
--

ALTER SEQUENCE credentials_id_seq OWNED BY credentials.id;


--
-- Name: databasechangelog; Type: TABLE; Schema: public; Owner: bartnet
--

CREATE TABLE databasechangelog (
    id character varying(255) NOT NULL,
    author character varying(255) NOT NULL,
    filename character varying(255) NOT NULL,
    dateexecuted timestamp with time zone NOT NULL,
    orderexecuted integer NOT NULL,
    exectype character varying(10) NOT NULL,
    md5sum character varying(35),
    description character varying(255),
    comments character varying(255),
    tag character varying(255),
    liquibase character varying(20)
);


--
-- Name: databasechangeloglock; Type: TABLE; Schema: public; Owner: bartnet
--

CREATE TABLE databasechangeloglock (
    id integer NOT NULL,
    locked boolean NOT NULL,
    lockgranted timestamp with time zone,
    lockedby character varying(255)
);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: bartnet
--

ALTER TABLE ONLY credentials ALTER COLUMN id SET DEFAULT nextval('credentials_id_seq'::regclass);


--
-- Name: pk_check_states; Type: CONSTRAINT; Schema: public; Owner: bartnet
--

ALTER TABLE ONLY check_states
    ADD CONSTRAINT pk_check_states PRIMARY KEY (check_id);


--
-- Name: pk_checks; Type: CONSTRAINT; Schema: public; Owner: bartnet
--

ALTER TABLE ONLY checks
    ADD CONSTRAINT pk_checks PRIMARY KEY (id);


--
-- Name: pk_credentials; Type: CONSTRAINT; Schema: public; Owner: bartnet
--

ALTER TABLE ONLY credentials
    ADD CONSTRAINT pk_credentials PRIMARY KEY (id);


--
-- Name: pk_databasechangeloglock; Type: CONSTRAINT; Schema: public; Owner: bartnet
--

ALTER TABLE ONLY databasechangeloglock
    ADD CONSTRAINT pk_databasechangeloglock PRIMARY KEY (id);


--
-- Name: bastion_id_idx; Type: INDEX; Schema: public; Owner: bartnet
--

CREATE INDEX bastion_id_idx ON check_state_memos USING btree (bastion_id);


--
-- Name: cust_execution_group_id_idx; Type: INDEX; Schema: public; Owner: bartnet
--

CREATE INDEX cust_execution_group_id_idx ON checks USING btree (customer_id, execution_group_id);


--
-- Name: execution_group_id_idx; Type: INDEX; Schema: public; Owner: bartnet
--

CREATE INDEX execution_group_id_idx ON checks USING btree (execution_group_id);


--
-- Name: idx_assertions_check_id_and_customer_id; Type: INDEX; Schema: public; Owner: bartnet
--

CREATE INDEX idx_assertions_check_id_and_customer_id ON assertions USING btree (check_id, customer_id);


--
-- Name: idx_checks_customer_id; Type: INDEX; Schema: public; Owner: bartnet
--

CREATE INDEX idx_checks_customer_id ON checks USING btree (customer_id);


--
-- Name: idx_credentials_customer_id; Type: INDEX; Schema: public; Owner: bartnet
--

CREATE INDEX idx_credentials_customer_id ON credentials USING btree (customer_id);


--
-- Name: idx_memos_bastion_id_check_id; Type: INDEX; Schema: public; Owner: bartnet
--

CREATE UNIQUE INDEX idx_memos_bastion_id_check_id ON check_state_memos USING btree (check_id, bastion_id);


--
-- Name: idx_memos_check_id; Type: INDEX; Schema: public; Owner: bartnet
--

CREATE INDEX idx_memos_check_id ON check_state_memos USING btree (check_id);


--
-- Name: update_checks; Type: TRIGGER; Schema: public; Owner: bartnet
--

CREATE TRIGGER update_checks BEFORE UPDATE ON checks FOR EACH ROW EXECUTE PROCEDURE update_time();


--
-- Name: update_credentials; Type: TRIGGER; Schema: public; Owner: bartnet
--

CREATE TRIGGER update_credentials BEFORE UPDATE ON credentials FOR EACH ROW EXECUTE PROCEDURE update_time();

