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

### Create Table `user_sessions`

    CREATE TABLE IF NOT EXISTS user_sessions (
        id INT AUTO_INCREMENT PRIMARY KEY,
        user_id BIGINT NOT NULL,
        refresh_token VARCHAR(36) NOT NULL,
        user_agent TEXT,
        revoked BIGINT NOT NULL DEFAULT 0,
        expired_at BIGINT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    );

    CREATE INDEX idx_users_id ON user_sessions(user_id);
    CREATE UNIQUE INDEX idx_refresh_token ON user_sessions(refresh_token);

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

## 🚀 Cara pakai HMAC di client pre request POSTMAN

```javascript

const secret = "12345678901234567890123456789012"; // samakan dengan backend

const method = pm.request.method;
const nonce = crypto.randomUUID();
const timestamp = Math.floor(Date.now() / 1000).toString();

let body = "";

if (pm.request.body && pm.request.body.raw) {
    body = pm.variables.replaceIn(pm.request.body.raw);
    const obj = JSON.parse(body);
    body = JSON.stringify(obj);

    // update body request supaya sama dengan yang disign
    pm.request.body.raw = body;
}

const payload = `${method}:${nonce}:${timestamp}:${body}`;

const signature = CryptoJS.HmacSHA256(payload, secret)
    .toString(CryptoJS.enc.Hex);

pm.request.headers.upsert({ key: "X-Nonce", value: nonce });
pm.request.headers.upsert({ key: "X-Timestamp", value: timestamp });
pm.request.headers.upsert({ key: "X-Signature", value: signature });

console.log("Nonce:", nonce)
console.log("timestamp:", timestamp)
console.log("Payload:", payload);
console.log("Signature:", signature);


```

## Build untuk linux
``` bash
GOOS=linux GOARCH=amd64 go build -o go_api main.go

```



