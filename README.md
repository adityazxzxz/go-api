# Go API

Simple REST API built with Go & MySQL.

---



### System Requirement
* Mysql v8.0.40  
* Go (1.25+ recommended)
* Redis 7.4.7
- Air (optional)

---

### Application Mode

    GIN_MODE=release/debug
release untuk production  
debug untuk development  
gin_mode dapat di set .env

---



## Database Setup (MySQL)

### Create Database

    CREATE DATABASE IF NOT EXISTS mydatabase
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

Gunakan database:

    USE mydatabase;

---

### Create Table `users`

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
        login_type VARCHAR(255),
        id_migration VARCHAR(255),
        migration_date datetime DEFAULT NULL,
        last_login BIGINT,
        last_ip VARCHAR(45),
        created_at BIGINT,
        updated_at BIGINT,
        deleted_at BIGINT
    );

---

### Recommended Indexes

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

### Create table `email_templates`

    CREATE TABLE email_templates (
        id BIGINT AUTO_INCREMENT PRIMARY KEY,
        uuid VARCHAR(36) NOT NULL,
        template_name VARCHAR(255) NOT NULL,
        body TEXT,
        prev_data TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
        deleted_at TIMESTAMP NULL DEFAULT NULL,
        created_by BIGINT NULL,
        updated_by BIGINT NULL,
        deleted_by BIGINT NULL,
        INDEX idx_template_name (template_name),
        UNIQUE KEY uniq_uuid (uuid));

---

### Insert mail example body
    INSERT INTO email_templates (
        uuid,
        template_name,
        body,
        prev_data,
        created_by
    ) VALUES (
        UUID(),
        'magic_link',
        '<h1>Halo $nama$</h1>
    <p>Kode kamu: <b>$kode$</b></p>',
        NULL,
        1
    );
---



## Running the App

### Normal Run

    go run main.go

---

### Dev Mode (Hot Reload)

Menggunakan Air:

    air

Install Air (jika belum):

    go install github.com/air-verse/air@latest

---

## Cara pakai HMAC di client pre request POSTMAN

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



