CREATE TABLE IF NOT EXISTS wishlist_items (
    customer_id TEXT        NOT NULL,
    product_id  TEXT        NOT NULL,
    added_at    TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (customer_id, product_id)
);