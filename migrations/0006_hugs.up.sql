--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

--
-- Name: notification_type; Type: TYPE; Schema: public; Owner: hugs
--

CREATE TYPE notification_type AS ENUM (
    'webhook',
    'slack_bot',
    'email',
    'pagerduty'
);


--ALTER TYPE notification_type OWNER TO hugs;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: default_notifications; Type: TABLE; Schema: public; Owner: hugs; Tablespace: 
--

CREATE TABLE default_notifications (
    id integer NOT NULL,
    customer_id uuid NOT NULL,
    type notification_type NOT NULL,
    value character varying(255) NOT NULL
);


--ALTER TABLE default_notifications OWNER TO hugs;

--
-- Name: default_notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: hugs
--

CREATE SEQUENCE default_notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--ALTER TABLE default_notifications_id_seq OWNER TO hugs;

--
-- Name: default_notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hugs
--

ALTER SEQUENCE default_notifications_id_seq OWNED BY default_notifications.id;


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: hugs; Tablespace: 
--

CREATE TABLE notifications (
    id integer NOT NULL,
    check_id character varying(255) NOT NULL,
    customer_id uuid NOT NULL,
    user_id integer NOT NULL,
    value character varying(255) NOT NULL,
    type notification_type
);


--ALTER TABLE notifications OWNER TO hugs;

--
-- Name: notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: hugs
--

CREATE SEQUENCE notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--ALTER TABLE notifications_id_seq OWNER TO hugs;

--
-- Name: notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hugs
--

ALTER SEQUENCE notifications_id_seq OWNED BY notifications.id;


--
-- Name: pagerduty_oauth_responses; Type: TABLE; Schema: public; Owner: hugs; Tablespace: 
--

CREATE TABLE pagerduty_oauth_responses (
    id integer NOT NULL,
    customer_id uuid NOT NULL,
    data jsonb NOT NULL
);


--ALTER TABLE pagerduty_oauth_responses OWNER TO hugs;

--
-- Name: pagerduty_oauth_responses_id_seq; Type: SEQUENCE; Schema: public; Owner: hugs
--

CREATE SEQUENCE pagerduty_oauth_responses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--ALTER TABLE pagerduty_oauth_responses_id_seq OWNER TO hugs;

--
-- Name: pagerduty_oauth_responses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hugs
--

ALTER SEQUENCE pagerduty_oauth_responses_id_seq OWNED BY pagerduty_oauth_responses.id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: hugs; Tablespace: 
-- NOTE this already exists
--

--CREATE TABLE schema_migrations (
--    version integer NOT NULL
--);


--ALTER TABLE schema_migrations OWNER TO hugs;

--
-- Name: slack_oauth_responses; Type: TABLE; Schema: public; Owner: hugs; Tablespace: 
--

CREATE TABLE slack_oauth_responses (
    id integer NOT NULL,
    customer_id uuid,
    data jsonb NOT NULL
);


--ALTER TABLE slack_oauth_responses OWNER TO hugs;

--
-- Name: slack_oauth_responses_id_seq; Type: SEQUENCE; Schema: public; Owner: hugs
--

CREATE SEQUENCE slack_oauth_responses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--ALTER TABLE slack_oauth_responses_id_seq OWNER TO hugs;

--
-- Name: slack_oauth_responses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: hugs
--

ALTER SEQUENCE slack_oauth_responses_id_seq OWNED BY slack_oauth_responses.id;


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: hugs
--

ALTER TABLE ONLY default_notifications ALTER COLUMN id SET DEFAULT nextval('default_notifications_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: hugs
--

ALTER TABLE ONLY notifications ALTER COLUMN id SET DEFAULT nextval('notifications_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: hugs
--

ALTER TABLE ONLY pagerduty_oauth_responses ALTER COLUMN id SET DEFAULT nextval('pagerduty_oauth_responses_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: hugs
--

ALTER TABLE ONLY slack_oauth_responses ALTER COLUMN id SET DEFAULT nextval('slack_oauth_responses_id_seq'::regclass);


--
-- Name: default_notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: hugs; Tablespace: 
--

ALTER TABLE ONLY default_notifications
    ADD CONSTRAINT default_notifications_pkey PRIMARY KEY (id);


--
-- Name: notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: hugs; Tablespace: 
--

ALTER TABLE ONLY notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: pagerduty_oauth_responses_pkey; Type: CONSTRAINT; Schema: public; Owner: hugs; Tablespace: 
--

ALTER TABLE ONLY pagerduty_oauth_responses
    ADD CONSTRAINT pagerduty_oauth_responses_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: hugs; Tablespace: 
--

--ALTER TABLE ONLY schema_migrations
--    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: slack_oauth_responses_pkey; Type: CONSTRAINT; Schema: public; Owner: hugs; Tablespace: 
--

ALTER TABLE ONLY slack_oauth_responses
    ADD CONSTRAINT slack_oauth_responses_pkey PRIMARY KEY (id);


--
-- Name: idx_default_notifications_customer; Type: INDEX; Schema: public; Owner: hugs; Tablespace: 
--

CREATE INDEX idx_default_notifications_customer ON default_notifications USING btree (customer_id);


--
-- Name: notifications_customer_id_check_id_idx; Type: INDEX; Schema: public; Owner: hugs; Tablespace: 
--

CREATE INDEX notifications_customer_id_check_id_idx ON notifications USING btree (customer_id, check_id);


--
-- Name: slack_oauth_responses_customer_id_idx; Type: INDEX; Schema: public; Owner: hugs; Tablespace: 
--

CREATE INDEX slack_oauth_responses_customer_id_idx ON slack_oauth_responses USING btree (customer_id);


--
-- Name: public; Type: ACL; Schema: -; Owner: hugs
--

--REVOKE ALL ON SCHEMA public FROM PUBLIC;
--REVOKE ALL ON SCHEMA public FROM hugs;
--GRANT ALL ON SCHEMA public TO hugs;
--GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

