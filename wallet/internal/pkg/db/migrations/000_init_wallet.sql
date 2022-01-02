CREATE TABLE IF NOT EXISTS wallets (
  id SERIAL PRIMARY KEY,
  user_id bigint NOT NULL CHECK (user_id > 0),
  balance decimal NOT NULL
);

CREATE TABLE IF NOT EXISTS wallet_transactions (
  id SERIAL PRIMARY KEY,
  wallet_id bigint REFERENCES wallets ON DELETE SET NULL,
  order_id bigint,
  cost decimal NOT NULL,
  type smallint NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO wallets(user_id, balance) VALUES(1, 100), (2, 100), (3, 100);
