-- Enable the uuid-ossp extension to generate UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create customers table with UUID as primary key and timestamps
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);


-- Create products table with UUID as primary key and timestamps
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    price DECIMAL NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Define the enum type for order status
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'order_status') THEN
        CREATE TYPE order_status AS ENUM ('pending', 'completed', 'failed');
    END IF;
END $$;

-- Create orders table with UUID as primary key and foreign keys to customers and products
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id UUID NOT NULL,
    product_id UUID NOT NULL,
    status order_status NOT NULL DEFAULT 'pending',
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

-- Insert initial data
INSERT INTO customers (id, name) VALUES (uuid_generate_v4(), 'John Doe') ON CONFLICT DO NOTHING;
INSERT INTO products (id, name, price) VALUES (uuid_generate_v4(), 'Sample Product', 99.99) ON CONFLICT DO NOTHING;