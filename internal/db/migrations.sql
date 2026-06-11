CREATE DATABASE IF NOT EXISTS biga CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE biga;

CREATE TABLE IF NOT EXISTS sim_account (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    name          VARCHAR(64)  NOT NULL DEFAULT 'default',
    cash          DECIMAL(18,2) NOT NULL,
    frozen_cash   DECIMAL(18,2) NOT NULL DEFAULT 0,
    created_at    DATETIME(3) NOT NULL,
    updated_at    DATETIME(3) NOT NULL,
    UNIQUE KEY uk_name (name)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS sim_position (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    account_id    BIGINT NOT NULL,
    code          CHAR(6) NOT NULL,
    name          VARCHAR(64) NOT NULL,
    board         VARCHAR(16) NOT NULL DEFAULT 'main',
    quantity      INT NOT NULL,
    available     INT NOT NULL,
    frozen_qty    INT NOT NULL DEFAULT 0,
    cost_price    DECIMAL(12,4) NOT NULL,
    buy_date      DATE NOT NULL,
    updated_at    DATETIME(3) NOT NULL,
    UNIQUE KEY uk_account_code (account_id, code),
    KEY idx_account (account_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS sim_order (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    account_id    BIGINT NOT NULL,
    code          CHAR(6) NOT NULL,
    name          VARCHAR(64) NOT NULL,
    side          ENUM('buy','sell') NOT NULL,
    order_type    ENUM('limit','market') NOT NULL,
    price         DECIMAL(12,4) NULL,
    quantity      INT NOT NULL,
    filled_qty    INT NOT NULL DEFAULT 0,
    status        ENUM('pending','partial','filled','cancelled','rejected') NOT NULL,
    reject_reason VARCHAR(255) NULL,
    source        VARCHAR(32) NOT NULL DEFAULT 'ai',
    remark        VARCHAR(255) NULL,
    created_at    DATETIME(3) NOT NULL,
    updated_at    DATETIME(3) NOT NULL,
    KEY idx_account_status (account_id, status),
    KEY idx_code (code)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS sim_trade (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    account_id    BIGINT NOT NULL,
    order_id      BIGINT NOT NULL,
    code          CHAR(6) NOT NULL,
    name          VARCHAR(64) NOT NULL,
    side          ENUM('buy','sell') NOT NULL,
    price         DECIMAL(12,4) NOT NULL,
    quantity      INT NOT NULL,
    amount        DECIMAL(18,2) NOT NULL,
    commission    DECIMAL(12,4) NOT NULL,
    stamp_tax     DECIMAL(12,4) NOT NULL,
    transfer_fee  DECIMAL(12,4) NOT NULL,
    trade_date    DATE NOT NULL,
    traded_at     DATETIME(3) NOT NULL,
    KEY idx_account_date (account_id, trade_date),
    KEY idx_order (order_id)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS market_quote_log (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    code          CHAR(6) NOT NULL,
    name          VARCHAR(64) NOT NULL,
    source        VARCHAR(16) NOT NULL,
    price         DECIMAL(12,4) NOT NULL,
    open_p        DECIMAL(12,4) NOT NULL,
    high_p        DECIMAL(12,4) NOT NULL,
    low_p         DECIMAL(12,4) NOT NULL,
    prev_close    DECIMAL(12,4) NOT NULL,
    change_pct    DECIMAL(10,4) NOT NULL,
    volume        BIGINT NOT NULL,
    amount        DECIMAL(20,2) NOT NULL,
    bid1          DECIMAL(12,4) NOT NULL,
    ask1          DECIMAL(12,4) NOT NULL,
    limit_up      DECIMAL(12,4) NOT NULL,
    limit_down    DECIMAL(12,4) NOT NULL,
    payload_json  JSON NOT NULL,
    recorded_at   DATETIME(3) NOT NULL,
    KEY idx_code_time (code, recorded_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS sim_settlement_log (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    account_id    BIGINT NOT NULL,
    trade_date    DATE NOT NULL,
    action        VARCHAR(32) NOT NULL,
    detail_json   JSON NULL,
    created_at    DATETIME(3) NOT NULL,
    KEY idx_account_date (account_id, trade_date)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS ai_action_log (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    account_id    BIGINT NOT NULL,
    action        VARCHAR(64) NOT NULL,
    request_json  JSON NULL,
    response_json JSON NULL,
    success       TINYINT(1) NOT NULL,
    error_msg     VARCHAR(512) NULL,
    created_at    DATETIME(3) NOT NULL,
    KEY idx_account_time (account_id, created_at)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS stock_info (
    code          CHAR(6) PRIMARY KEY,
    name          VARCHAR(64) NOT NULL,
    board         VARCHAR(16) NOT NULL,
    updated_at    DATETIME(3) NOT NULL
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS market_kline (
    id            BIGINT PRIMARY KEY AUTO_INCREMENT,
    code          CHAR(6) NOT NULL,
    period        VARCHAR(16) NOT NULL,
    trade_date    DATE NOT NULL,
    open_p        DECIMAL(12,4) NOT NULL,
    high_p        DECIMAL(12,4) NOT NULL,
    low_p         DECIMAL(12,4) NOT NULL,
    close_p       DECIMAL(12,4) NOT NULL,
    volume        BIGINT NOT NULL,
    amount        DECIMAL(20,2) NOT NULL DEFAULT 0,
    source        VARCHAR(16) NOT NULL,
    recorded_at   DATETIME(3) NOT NULL,
    UNIQUE KEY uk_code_period_date (code, period, trade_date),
    KEY idx_code_period (code, period, trade_date)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS sim_conditional_order (
    id                BIGINT PRIMARY KEY AUTO_INCREMENT,
    account_id        BIGINT NOT NULL,
    code              CHAR(6) NOT NULL,
    name              VARCHAR(64) NOT NULL,
    condition_type    VARCHAR(32) NOT NULL,
    trigger_value     DECIMAL(12,4) NOT NULL,
    side              ENUM('buy','sell') NOT NULL,
    order_type        ENUM('limit','market') NOT NULL,
    action_price      DECIMAL(12,4) NULL,
    quantity          INT NOT NULL,
    status            ENUM('pending','triggering','triggered','failed','cancelled','expired') NOT NULL DEFAULT 'pending',
    trigger_order_id  BIGINT NULL,
    trigger_message   VARCHAR(255) NULL,
    valid_date        DATE NOT NULL,
    source            VARCHAR(32) NOT NULL DEFAULT 'api',
    remark            VARCHAR(255) NULL,
    created_at        DATETIME(3) NOT NULL,
    updated_at        DATETIME(3) NOT NULL,
    triggered_at      DATETIME(3) NULL,
    KEY idx_account_status (account_id, status),
    KEY idx_pending_date (status, valid_date),
    KEY idx_code (code)
) ENGINE=InnoDB;
