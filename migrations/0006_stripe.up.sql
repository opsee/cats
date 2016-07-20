-- insert into customers (name) values ('mark');
-- insert into customers (name) values ('mark2');

CREATE TABLE subscriptions (
    id serial primary key,
    plan character varying(64) not null default 'free',
    stripe_customer_id character varying(64),
    stripe_subscription_id character varying(64),
    quantity integer not null default 0,
    status character varying(32) not null default 'trialing',
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);

alter table customers drop column subscription;
drop type subscription_types;
alter table customers add column subscription_id integer unique references subscriptions (id) on delete cascade;
create index idx_subscriptions_stripe_customer_id ON subscriptions USING btree (stripe_customer_id);
create index idx_subscriptions_stripe_subscription_id ON subscriptions USING btree (stripe_subscription_id);
create trigger update_subscriptions before update on subscriptions for each row execute procedure update_time();

with custs as (
  select id, nextval('subscriptions_id_seq') as sub_id from customers
),
subs as (
  insert into subscriptions (id) select sub_id from custs
)
update customers as c set subscription_id = custs.sub_id from custs where c.id = custs.id;

alter table customers alter column subscription_id set not null;
