-- Create a sample database
-- Note: We use a psql conditional to check if database exists
SELECT 'CREATE DATABASE sampledb'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'sampledb')\gexec

-- Connect to the sample database
\c sampledb

-- Create a users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    full_name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create a posts table
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    content TEXT,
    published BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample users
INSERT INTO users (username, email, full_name) VALUES
    ('alice', 'alice@example.com', 'Alice Smith'),
    ('bob', 'bob@example.com', 'Bob Johnson'),
    ('charlie', 'charlie@example.com', 'Charlie Brown'),
    ('diana', 'diana@example.com', 'Diana Prince'),
    ('eve', 'eve@example.com', 'Eve Anderson')
ON CONFLICT (username) DO NOTHING;

-- Insert sample posts
INSERT INTO posts (user_id, title, content, published) VALUES
    (1, 'Getting Started with PostgreSQL', 'PostgreSQL is a powerful, open source object-relational database system...', true),
    (1, 'Advanced SQL Queries', 'Learn how to write complex SQL queries with joins and subqueries...', true),
    (2, 'Introduction to Nix', 'Nix is a powerful package manager for Linux and other Unix systems...', true),
    (3, 'Building Web Applications', 'A comprehensive guide to building modern web applications...', false),
    (4, 'Database Design Best Practices', 'Follow these best practices when designing your database schema...', true),
    (5, 'Draft: Future of Databases', 'This is a draft post about emerging database technologies...', false)
ON CONFLICT DO NOTHING;

-- Display the data
\echo ''
\echo '========================================='
\echo 'Sample data initialized successfully!'
\echo '========================================='
\echo ''
\echo 'Users table:'
SELECT * FROM users;
\echo ''
\echo 'Posts table:'
SELECT id, user_id, title, published, created_at FROM posts;
\echo ''
