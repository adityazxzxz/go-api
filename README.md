# Go API

Simple REST API built with Go & MySQL.

---

## 📦 Database Setup (MySQL)

### 1️⃣ Create Database

    CREATE DATABASE IF NOT EXISTS mydatabase
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

Gunakan database:

    USE mydatabase;

---

### 2️⃣ Create Table `users`

    CREATE TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        uuid VARCHAR(36) NOT NULL,
        first_name VARCHAR(100),
        last_name VARCHAR(100),
        username VARCHAR(100),
        email VARCHAR(150),
        phone VARCHAR(30),
        status INT DEFAULT 0,
        password VARCHAR(255),
        last_login BIGINT,
        last_ip VARCHAR(45),
        created_at BIGINT,
        updated_at BIGINT,
        deleted_at BIGINT
    );

---

### 3️⃣ Recommended Indexes

    CREATE UNIQUE INDEX idx_users_uuid ON users(uuid);
    CREATE UNIQUE INDEX idx_users_username ON users(username);
    CREATE UNIQUE INDEX idx_users_email ON users(email);

---



## 🚀 Running the App

### ▶ Normal Run

    go run main.go

---

### 🔥 Dev Mode (Hot Reload)

Menggunakan Air:

    air

Install Air (jika belum):

    go install github.com/air-verse/air@latest

---

## ✅ Requirements

- Go (1.20+ recommended)
- MySQL
- Air (optional)

---

