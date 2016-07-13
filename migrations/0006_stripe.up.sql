alter table customers add column stripe_customer_id character varying(64);
alter table customers add column stripe_subscription_id character varying(64);
alter table customers add column subscription_quantity integer not null default 0;
create index idx_customers_stripe_customer_id ON customers USING btree (stripe_customer_id);
create index idx_customers_stripe_subscription_id ON customers USING btree (stripe_subscription_id);
