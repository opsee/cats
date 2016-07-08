CREATE TABLE check_state_transitions (
    id integer PRIMARY KEY,
    check_id character varying(255) NOT NULL REFERENCES checks (id),
    customer_id uuid NOT NULL,
    from_state integer NOT NULL,
    to_state integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE SEQUENCE check_state_transitions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
    OWNED BY check_state_transitions.id;

ALTER TABLE ONLY check_state_transitions ALTER COLUMN id SET DEFAULT nextval('check_state_transitions_id_seq'::regclass);

CREATE INDEX idx_check_state_transitions_check_id ON check_state_transitions (check_id);