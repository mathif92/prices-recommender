-- +goose Up
CREATE TABLE IF NOT EXISTS "hotels" (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    link TEXT NOT NULL,
    rating DECIMAL(3, 2) NOT NULL,
    location VARCHAR(255) NOT NULL,
    is_all_inclusive BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    latitude_coordinate DECIMAL(10, 7) DEFAULT 0,
    longitude_coordinate DECIMAL(10, 7) DEFAULT 0,
    class INTEGER DEFAULT 0,
    overall_rating DECIMAL(3, 2) DEFAULT 0,
    CONSTRAINT hotels_name_location_unique UNIQUE (name, location)
);

CREATE TABLE IF NOT EXISTS "hotel_ratings" (
    id SERIAL PRIMARY KEY,
    hotel_id INTEGER NOT NULL,
    rating DECIMAL(3, 2) NOT NULL,
    review_count INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT hotel_ratings_hotel_id_rating_unique UNIQUE (hotel_id, rating),
    FOREIGN KEY (hotel_id) REFERENCES hotels(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "hotel_reviews" (
    id SERIAL PRIMARY KEY,
    hotel_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    total_mentioned INTEGER DEFAULT 0,
    positive INTEGER NOT NULL,
    negative INTEGER NOT NULL,
    neutral INTEGER NOT NULL,
    external_link TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT hotel_reviews_hotel_id_name_unique UNIQUE (hotel_id, name),
    FOREIGN KEY (hotel_id) REFERENCES hotels(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS "prices" (
    id SERIAL PRIMARY KEY,
    hotel_id INTEGER NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    currency VARCHAR(10) NOT NULL,
    checkin_time TEXT NULL,
    checkout_time TEXT NULL,
    property_details_link TEXT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT prices_hotel_id_start_date_end_date_unique UNIQUE (hotel_id, start_date, end_date),
    FOREIGN KEY (hotel_id) REFERENCES hotels(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_hotels_name ON hotels(name);
CREATE INDEX IF NOT EXISTS idx_prices_hotel_id ON prices(hotel_id);
CREATE INDEX IF NOT EXISTS idx_prices_start_date_end_date ON prices(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_prices_price ON prices(price);



-- +goose Down
DROP TABLE IF EXISTS "prices";
DROP TABLE IF EXISTS "hotel_reviews";
DROP TABLE IF EXISTS "hotel_ratings";
DROP TABLE IF EXISTS "hotels";