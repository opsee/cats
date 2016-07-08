--
-- PostgreSQL database dump
--

-- Dumped from database version 9.4.5
-- Dumped by pg_dump version 9.5.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;

SET search_path = public, pg_catalog;

--
-- Name: status_types; Type: TYPE; Schema: public; Owner: vape
--

CREATE TYPE status_types AS ENUM (
    'invited',
    'active',
    'inactive'
);


--
-- Name: subscription_types; Type: TYPE; Schema: public; Owner: vape
--

CREATE TYPE subscription_types AS ENUM (
    'free',
    'basic',
    'advanced'
);


--
-- Name: insert_userdata(); Type: FUNCTION; Schema: public; Owner: vape
--

CREATE FUNCTION insert_userdata() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
	BEGIN
	insert into userdata (user_id) values(NEW.id);
	RETURN NEW;
  END;
$$;


--
-- Name: json_merge(jsonb, jsonb); Type: FUNCTION; Schema: public; Owner: vape
--

CREATE FUNCTION json_merge(data jsonb, merge_data jsonb) RETURNS jsonb
    LANGUAGE sql IMMUTABLE
    AS $$
    SELECT ('{'||string_agg(to_json(key)||':'||value, ',')||'}')::jsonb
    FROM (
        WITH to_merge AS (
            SELECT * FROM jsonb_each(merge_data)
        )
        SELECT *
        FROM jsonb_each(data)
        WHERE key NOT IN (SELECT key FROM to_merge)
        UNION ALL
        SELECT * FROM to_merge
    ) t;
$$;


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: customers; Type: TABLE; Schema: public; Owner: vape
--

CREATE TABLE customers (
    id uuid DEFAULT uuid_generate_v1mc() NOT NULL,
    name character varying(255) DEFAULT 'default'::character varying,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    active boolean DEFAULT false NOT NULL,
    subscription subscription_types DEFAULT 'free'::subscription_types NOT NULL
);


--
-- Name: signups; Type: TABLE; Schema: public; Owner: vape
--

CREATE TABLE signups (
    id integer NOT NULL,
    email character varying(255),
    name character varying(255),
    claimed boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    activated boolean DEFAULT false NOT NULL,
    referrer character varying(255) DEFAULT ''::character varying NOT NULL,
    customer_id character varying(36) DEFAULT ''::character varying NOT NULL,
    perms bigint DEFAULT 0 NOT NULL
);


--
-- Name: signups_id_seq; Type: SEQUENCE; Schema: public; Owner: vape
--

CREATE SEQUENCE signups_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: signups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: vape
--

ALTER SEQUENCE signups_id_seq OWNED BY signups.id;


--
-- Name: userdata; Type: TABLE; Schema: public; Owner: vape
--

CREATE TABLE userdata (
    user_id integer NOT NULL,
    data jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: vape
--

CREATE TABLE users (
    id integer NOT NULL,
    email character varying(255) NOT NULL,
    password_hash character varying(60) NOT NULL,
    admin boolean DEFAULT false NOT NULL,
    active boolean DEFAULT false NOT NULL,
    verified boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    name character varying(255) NOT NULL,
    customer_id uuid,
    status status_types DEFAULT 'invited'::status_types NOT NULL,
    perms bigint DEFAULT 0 NOT NULL
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: vape
--

CREATE SEQUENCE users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: vape
--

ALTER SEQUENCE users_id_seq OWNED BY users.id;


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: vape
--

ALTER TABLE ONLY signups ALTER COLUMN id SET DEFAULT nextval('signups_id_seq'::regclass);


--
-- Name: id; Type: DEFAULT; Schema: public; Owner: vape
--

ALTER TABLE ONLY users ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);


--
-- Name: customers_pkey; Type: CONSTRAINT; Schema: public; Owner: vape
--

ALTER TABLE ONLY customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);


--
-- Name: pk_signups; Type: CONSTRAINT; Schema: public; Owner: vape
--

ALTER TABLE ONLY signups
    ADD CONSTRAINT pk_signups PRIMARY KEY (id);


--
-- Name: pk_users; Type: CONSTRAINT; Schema: public; Owner: vape
--

ALTER TABLE ONLY users
    ADD CONSTRAINT pk_users PRIMARY KEY (id);


--
-- Name: signups_email_key; Type: CONSTRAINT; Schema: public; Owner: vape
--

ALTER TABLE ONLY signups
    ADD CONSTRAINT signups_email_key UNIQUE (email);


--
-- Name: users_email_key; Type: CONSTRAINT; Schema: public; Owner: vape
--

ALTER TABLE ONLY users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: idx_signups_email; Type: INDEX; Schema: public; Owner: vape
--

CREATE UNIQUE INDEX idx_signups_email ON signups USING btree (lower((email)::text) varchar_pattern_ops);


--
-- Name: idx_signups_referrer; Type: INDEX; Schema: public; Owner: vape
--

CREATE INDEX idx_signups_referrer ON signups USING btree (referrer);


--
-- Name: idx_users_customers; Type: INDEX; Schema: public; Owner: vape
--

CREATE INDEX idx_users_customers ON users USING btree (customer_id);


--
-- Name: idx_users_email_active; Type: INDEX; Schema: public; Owner: vape
--

CREATE UNIQUE INDEX idx_users_email_active ON users USING btree (lower((email)::text) varchar_pattern_ops, active);


--
-- Name: insert_userdata; Type: TRIGGER; Schema: public; Owner: vape
--

CREATE TRIGGER insert_userdata AFTER INSERT ON users FOR EACH ROW EXECUTE PROCEDURE insert_userdata();


--
-- Name: update_signups; Type: TRIGGER; Schema: public; Owner: vape
--

CREATE TRIGGER update_signups BEFORE UPDATE ON signups FOR EACH ROW EXECUTE PROCEDURE update_time();


--
-- Name: update_userdata; Type: TRIGGER; Schema: public; Owner: vape
--

CREATE TRIGGER update_userdata BEFORE UPDATE ON userdata FOR EACH ROW EXECUTE PROCEDURE update_time();


--
-- Name: update_users; Type: TRIGGER; Schema: public; Owner: vape
--

CREATE TRIGGER update_users BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE update_time();


--
-- Name: fk_users_customers; Type: FK CONSTRAINT; Schema: public; Owner: vape
--

ALTER TABLE ONLY users
    ADD CONSTRAINT fk_users_customers FOREIGN KEY (customer_id) REFERENCES customers(id);


--
-- Name: userdata_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: vape
--

ALTER TABLE ONLY userdata
    ADD CONSTRAINT userdata_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

