CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INT, 
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id)
    REFERENCES users(id)

);