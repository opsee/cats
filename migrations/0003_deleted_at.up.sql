alter table checks add column deleted boolean default false not null;
create index idx_checks_customer_id_deleted ON checks USING btree (customer_id, deleted);
