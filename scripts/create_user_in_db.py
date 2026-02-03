#!/usr/bin/env python3
"""Manually create user in user-service database"""

import psycopg2
import hashlib

# Database connection
conn = psycopg2.connect(
    host="localhost",
    port=5433,
    user="splitwise",
    password="splitwise",
    dbname="splitwise"
)

cursor = conn.cursor()

# User data (matching Keycloak)
user_id = "a9cdabff-b34f-40a5-a1f2-0262755aa7cb"
name = "Vinay Gupta"
email = "vinay@email.com"
password_hash = "$2a$10$dummy_hash_placeholder"  # Not used since auth is via Keycloak

# Insert user
cursor.execute("""
    INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
    VALUES (%s, %s, %s, %s, NOW(), NOW())
    ON CONFLICT (id) DO UPDATE 
    SET name = EXCLUDED.name, email = EXCLUDED.email
""", (user_id, name, email, password_hash))

conn.commit()
print(f"✅ User {email} (ID: {user_id}) created/updated in database!")

cursor.close()
conn.close()
