USE ctf_manager_db;

CREATE TABLE IF NOT EXISTS instances (
    instance_id VARCHAR(255) PRIMARY KEY,
    image_tag VARCHAR(255) NOT NULL,
    runner_url VARCHAR(255) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL,
    state VARCHAR(50) NOT NULL,
    ttl_seconds BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    INDEX idx_runner_url (runner_url),
    INDEX idx_state (state),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
