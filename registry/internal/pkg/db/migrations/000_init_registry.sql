CREATE TABLE IF NOT EXISTS orders (
  id SERIAL PRIMARY KEY,
  user_id bigint NOT NULL CHECK (user_id > 0),
  status smallint NOT NULL DEFAULT 0,
  rejected_reason smallint NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
  id SERIAL PRIMARY KEY,
  title varchar(255) NOT NULL,
  price decimal(3) NOT NULL
);

CREATE TABLE IF NOT EXISTS order_items (
  id SERIAL PRIMARY KEY,
  order_id bigint NOT NULL REFERENCES orders ON DELETE CASCADE,
  product_id bigint REFERENCES products ON DELETE SET NULL,
  count smallint NOT NULL CHECK (count > 0),
  product_price decimal(3) NOT NULL
);

INSERT INTO products(title, price) VALUES ('A', 1.0), ('B', 2.0), ('C', 3.0);