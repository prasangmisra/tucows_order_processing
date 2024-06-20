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
	customer_id INT NOT NULL REFERENCES customers(id),
	product_id INT NOT NULL REFERENCES products(id),
	status TEXT NOT NULL,
	amount DECIMAL NOT NULL
	);

INSERT INTO customers (id, name) VALUES (1, 'John Doe') ON CONFLICT DO NOTHING;
INSERT INTO products (id, name, price) VALUES (1, 'Sample Product', 99.99) ON CONFLICT DO NOTHING;