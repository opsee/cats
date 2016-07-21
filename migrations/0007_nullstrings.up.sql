update subscriptions set stripe_subscription_id = '' where stripe_subscription_id is null;
update subscriptions set stripe_customer_id = '' where stripe_customer_id is null;
alter table subscriptions alter column stripe_subscription_id set not null;
alter table subscriptions alter column stripe_subscription_id set default '';
alter table subscriptions alter column stripe_customer_id set not null;
alter table subscriptions alter column stripe_customer_id set default '';
  