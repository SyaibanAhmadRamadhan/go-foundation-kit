CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL
);
INSERT INTO users (name)
SELECT 'user-' || g
FROM generate_series(1, 10000) AS g;
-- sebuah query yang akan sering dipakai di load test
CREATE INDEX IF NOT EXISTS idx_users_name ON users (name);
-- Customers (1M)
CREATE TABLE IF NOT EXISTS customers (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Products (200k) dengan teks & jsonb untuk FTS/GIN
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    sku TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    attrs JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Categories (300) + bridge
CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS product_categories (
    product_id BIGINT NOT NULL REFERENCES products(id),
    category_id BIGINT NOT NULL REFERENCES categories(id),
    PRIMARY KEY (product_id, category_id)
);
-- Orders (5M)
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL REFERENCES customers(id),
    status TEXT NOT NULL,
    -- e.g. 'paid','shipped','cancelled'
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Order items (20M)
CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id),
    product_id BIGINT NOT NULL REFERENCES products(id),
    quantity INT NOT NULL,
    unit_price_cents INT NOT NULL
);
-- Reviews (8M) untuk window & sentiment dummy
CREATE TABLE IF NOT EXISTS reviews (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id),
    customer_id BIGINT NOT NULL REFERENCES customers(id),
    rating INT NOT NULL CHECK (
        rating BETWEEN 1 AND 5
    ),
    comment TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Event log (jsonb, 10M)
CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL REFERENCES customers(id),
    kind TEXT NOT NULL,
    -- 'view','click','cart','purchase'
    payload JSONB NOT NULL,
    -- arbitrary
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- Indexes (kritikal)
CREATE INDEX IF NOT EXISTS idx_orders_customer_created ON orders(customer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product ON order_items(product_id);
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_products_created ON products(created_at);
CREATE INDEX IF NOT EXISTS idx_reviews_prod_created ON reviews(product_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_customer_created ON events(customer_id, created_at DESC);
-- FTS + GIN
ALTER TABLE products
ADD COLUMN IF NOT EXISTS tsv tsvector;
CREATE INDEX IF NOT EXISTS idx_products_tsv ON products USING GIN(tsv);
UPDATE products
SET tsv = to_tsvector(
        'simple',
        coalesce(title, '') || ' ' || coalesce(description, '')
    );
CREATE OR REPLACE FUNCTION products_tsv_trigger() RETURNS trigger AS $$ BEGIN NEW.tsv := to_tsvector(
        'simple',
        coalesce(NEW.title, '') || ' ' || coalesce(NEW.description, '')
    );
RETURN NEW;
END $$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS trg_products_tsv ON products;
CREATE TRIGGER trg_products_tsv BEFORE
INSERT
    OR
UPDATE ON products FOR EACH ROW EXECUTE FUNCTION products_tsv_trigger();
-- JSONB GIN (attrs & events.payload)
CREATE INDEX IF NOT EXISTS idx_products_attrs_gin ON products USING GIN(attrs jsonb_path_ops);
CREATE INDEX IF NOT EXISTS idx_events_payload_gin ON events USING GIN(payload jsonb_path_ops);
-- Seed contoh (pakai generate_series biar cepat; sesuaikan skala)
-- Categories
INSERT INTO categories(name)
SELECT 'Category ' || g
FROM generate_series(1, 300) g ON CONFLICT DO NOTHING;
-- Customers (1M) – bisa turunin kalau mesin kecil
INSERT INTO customers(name, email, created_at)
SELECT 'Customer ' || g,
    'cust' || g || '@ex.com',
    now() - (random() * 365 || ' days')::interval
FROM generate_series(1, 1000000) g;
-- Products (200k)
INSERT INTO products(sku, title, description, attrs, created_at)
SELECT 'SKU-' || g,
    'Product ' || g,
    'Great product number ' || g || ' with awesome features',
    jsonb_build_object(
        'brand',
        ('brand' ||(1 + floor(random() * 50))::int),
        'color',
        (ARRAY ['red','blue','green','black','white']) [1+floor(random()*5)],
        'tags',
        to_jsonb(
            ARRAY ['eco','new','sale','popular'] [1+floor(random()*4)]
        )
    ),
    now() - (random() * 365 || ' days')::interval
FROM generate_series(1, 200000) g;
-- Product-Categories (≈ avg 2/category per product)
INSERT INTO product_categories(product_id, category_id)
SELECT p.id,
    1 + (random() * 299)::int
FROM products p,
    generate_series(1, 2);
-- Orders (5M)
INSERT INTO orders(customer_id, status, created_at)
SELECT (1 + floor(random() * 1000000))::bigint,
    (ARRAY ['paid','shipped','cancelled']) [1+floor(random()*3)],
    now() - (random() * 365 || ' days')::interval
FROM generate_series(1, 5000000) g;
-- Order items (20M)
INSERT INTO order_items(order_id, product_id, quantity, unit_price_cents)
SELECT o.id,
    1 + (random() * 199999)::bigint,
    1 + (random() * 4)::int,
    500 + (random() * 50000)::int
FROM orders o,
    LATERAL generate_series(1, (1 + (random() * 3)::int));
-- Reviews (8M)
INSERT INTO reviews(
        product_id,
        customer_id,
        rating,
        comment,
        created_at
    )
SELECT 1 + (random() * 199999)::bigint,
    1 + (random() * 1000000)::bigint,
    1 + (random() * 5)::int,
    'lorem ipsum ' || g,
    now() - (random() * 365 || ' days')::interval
FROM generate_series(1, 8000000) g;
-- Events (10M)
INSERT INTO events(customer_id, kind, payload, created_at)
SELECT 1 + (random() * 1000000)::bigint,
    (ARRAY ['view','click','cart','purchase']) [1+floor(random()*4)],
    jsonb_build_object(
        'path',
        '/p/' ||(1 + (random() * 199999)::int),
        'ref',
        (ARRAY ['soc','ads','seo','direct']) [1+floor(random()*4)],
        'score',
        (random() * 100)::int
    ),
    now() - (random() * 90 || ' days')::interval
FROM generate_series(1, 10000000) g;