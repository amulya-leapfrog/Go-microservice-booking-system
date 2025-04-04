-- +goose Up
-- +goose StatementBegin
CREATE TABLE reservations (
    id SERIAL PRIMARY KEY,
    restaurant_id VARCHAR(255),
    user_id INT,
    count VARCHAR(50),
    reservation_time TIMESTAMPTZ,
    remarks TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE reservations;
-- +goose StatementEnd
