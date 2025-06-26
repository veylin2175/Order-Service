CREATE TABLE IF NOT EXISTS orders (
    order_uid          VARCHAR(255) PRIMARY KEY,
    track_number       VARCHAR(255),
    entry              VARCHAR(255),
    locale             VARCHAR(10),
    internal_signature VARCHAR(255),
    customer_id        VARCHAR(255),
    delivery_service   VARCHAR(255),
    shardkey           VARCHAR(255),
    sm_id              INTEGER,
    date_created       TIMESTAMP WITH TIME ZONE,
    oof_shard         VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS deliveries (
    order_uid VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid),
    name      VARCHAR(255),
    phone     VARCHAR(255),
    zip       VARCHAR(255),
    city      VARCHAR(255),
    address   VARCHAR(255),
    region    VARCHAR(255),
    email     VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS payments (
    order_uid      VARCHAR(255) PRIMARY KEY REFERENCES orders(order_uid),
    transaction    VARCHAR(255),
    request_id     VARCHAR(255),
    currency       VARCHAR(10),
    provider       VARCHAR(255),
    amount         INTEGER,
    payment_dt     BIGINT,
    bank           VARCHAR(255),
    delivery_cost  INTEGER,
    goods_total    INTEGER,
    custom_fee     INTEGER
);

CREATE TABLE IF NOT EXISTS items (
    id           SERIAL PRIMARY KEY,
    order_uid    VARCHAR(255) REFERENCES orders(order_uid),
    chrt_id      BIGINT,
    track_number VARCHAR(255),
    price        INTEGER,
    rid          VARCHAR(255),
    name         VARCHAR(255),
    sale         INTEGER,
    size         VARCHAR(255),
    total_price  INTEGER,
    nm_id        BIGINT,
    brand        VARCHAR(255),
    status       INTEGER
);