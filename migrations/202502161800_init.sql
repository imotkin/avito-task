-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS shop;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS shop.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username TEXT NOT NULL,
    hashed_password TEXT NOT NULL,
    coins INT NOT NULL DEFAULT 1000
);

CREATE TABLE IF NOT EXISTS shop.products (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    price INT NOT NULL
);

INSERT INTO shop.products (title, price)
VALUES ('t-shirt', 80),
       ('cup', 20),
       ('book', 50),
       ('pen', 10),
       ('powerbank', 200),
       ('hoody', 300),
       ('umbrella', 200),
       ('socks', 10),
       ('wallet', 50),
       ('pink-hoody', 500);

CREATE TABLE IF NOT EXISTS shop.inventory (
    holder UUID REFERENCES shop.users(id),
    product INT REFERENCES shop.products(id),
    amount INT NOT NULL DEFAULT 1,
    PRIMARY KEY (holder, product)
);

CREATE TABLE IF NOT EXISTS shop.transfers (
     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
     sender UUID NOT NULL REFERENCES shop.users(id),
     receiver UUID NOT NULL REFERENCES shop.users(id),
     amount INT NOT NULL
);

CREATE INDEX idx_users_name ON shop.users(username);
CREATE INDEX idx_transfers_receiver ON shop.transfers(receiver);
CREATE INDEX idx_transfers_sender ON shop.transfers(sender);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP SCHEMA IF EXISTS shop CASCADE;

-- +goose StatementEnd
