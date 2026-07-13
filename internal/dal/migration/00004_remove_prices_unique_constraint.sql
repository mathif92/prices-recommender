-- +goose Up
ALTER TABLE prices DROP CONSTRAINT IF EXISTS prices_hotel_id_start_date_end_date_unique;

-- +goose Down
ALTER TABLE prices ADD CONSTRAINT prices_hotel_id_start_date_end_date_unique UNIQUE (hotel_id, start_date, end_date);
