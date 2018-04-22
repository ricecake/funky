--
-- PostgreSQL database dump
--

-- Dumped from database version 10.1
-- Dumped by pg_dump version 10.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
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
-- Name: citext; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;


--
-- Name: EXTENSION citext; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION citext IS 'data type for case-insensitive character strings';


--
-- Name: ltree; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS ltree WITH SCHEMA public;


--
-- Name: EXTENSION ltree; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION ltree IS 'data type for hierarchical tree-like structures';


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: storage; Type: TABLE; Schema: public; Owner: guest
--

CREATE TABLE storage (
    id uuid DEFAULT uuid_generate_v4() NOT NULL,
    owner ltree NOT NULL,
    scope ltree NOT NULL,
    type citext NOT NULL,
    key citext NOT NULL,
    value json NOT NULL
);


ALTER TABLE storage OWNER TO guest;

--
-- Data for Name: storage; Type: TABLE DATA; Schema: public; Owner: guest
--

COPY storage (id, owner, scope, type, key, value) FROM stdin;
9ac60667-00ea-4347-bd64-8c80bab672fc	funky	universal	domain_route	localhost:8080	{"/ping":["test.ping","test.example"]}
3a8d14ca-7d13-43cb-8b05-94af7ceaf80d	funky	universal	event_route	test.ping	{"script":"example"}
1807af1d-e95e-47c8-9d9b-3144c715162e	funky	universal	event_route	test.example	{"script":"otherExample"}
\.


--
-- PostgreSQL database dump complete
--

