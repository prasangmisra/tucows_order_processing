CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
	name TEXT NOT NULL
    );

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	price DECIMAL NOT NULL
	);

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    customer_id INT NOT NULL,
    product_id INT NOT NULL,
    status VARCHAR(20) NOT NULL,
    amount FLOAT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_customer
        FOREIGN KEY(customer_id) 
        REFERENCES customers(id),
    CONSTRAINT fk_product
        FOREIGN KEY(product_id) 
        REFERENCES products(id)
);

INSERT INTO customers (id, name) VALUES (1, 'John Doe') ON CONFLICT DO NOTHING;
INSERT INTO products (id, name, price) VALUES (1, 'Sample Product', 99.99) ON CONFLICT DO NOTHING;