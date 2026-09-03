-- Create users table matching Spring Boot schema
CREATE TABLE IF NOT EXISTS users (
  username TEXT PRIMARY KEY,
  password TEXT NOT NULL,
  email TEXT,
  role TEXT,
  adGroup TEXT,
  createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
