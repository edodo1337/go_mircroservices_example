CREATE TABLE IF NOT EXISTS storage_items (
  id SERIAL PRIMARY KEY,
  product_id bigint NOT NULL CHECK (product_id > 0),
  count int NOT NULL CHECK (count > 0)
);

CREATE TABLE IF NOT EXISTS storage_transactions (
  id SERIAL PRIMARY KEY,
  order_id bigint NOT NULL CHECK (order_id > 0),
  type smallint NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE IF NOT EXISTS storage_transaction_items (
  id SERIAL PRIMARY KEY,
  order_id bigint NOT NULL CHECK (order_id > 0),
  product_id bigint NOT NULL CHECK (product_id > 0),
  transaction_id bigint NOT NULL REFERENCES wallets ON DELETE CASCADE,
  count int NOT NULL CHECK (count > 0)
);


INSERT INTO storage_items(product_id, count)
VALUES (1, 10), (2, 10), (3, 10); 
-- 