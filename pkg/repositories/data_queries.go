package repositories

type dbQuery int

const (
	insertHotels dbQuery = iota
	insertHotelRatings
	insertHotelReviews
	insertPrices
	insertUser
	insertUserSettings
	insertVacations
	listVacations
	deleteVacationsByYear
	listHotels
	getHotelByID
	listHotelsByLocation
	listHotelRatings
	listHotelReviews
	listPrices
	listHotelsPaginated
	countHotels
	listHotelsByLocationPaginated
	countHotelsByLocation
	listHotelsWithPrices
	listHotelsWithPricesPaginated
	countHotelsWithPrices
	listHotelsWithPricesByLocationPaginated
	countHotelsWithPricesByLocation
	listUserSettings
	getUserSetting
	deleteUserSetting
	getCurrentPrice
	getUserByEmail
	getUserByID
	getUserByGoogleID
	updateUser
)

var dbQueries map[dbQuery]string = map[dbQuery]string{
	insertVacations: `
		INSERT INTO vacations (name, start_date, end_date, affected_levels, year, created_at)
		VALUES (:name, :start_date, :end_date, :affected_levels, :year, :created_at)
		ON CONFLICT (name, year) DO UPDATE SET
			start_date = EXCLUDED.start_date,
			end_date = EXCLUDED.end_date,
			affected_levels = EXCLUDED.affected_levels,
			created_at = EXCLUDED.created_at
	`,
	listVacations: `
		SELECT id, name, start_date, end_date, affected_levels, year, created_at
		FROM vacations
		WHERE year = $1
		ORDER BY start_date
	`,
	deleteVacationsByYear: `
		DELETE FROM vacations WHERE year = $1
	`,
	
	insertHotels: `
		INSERT INTO hotels (name, description, link, rating, location, is_all_inclusive, created_at)
		VALUES (:name, :description, :link, :rating, :location, :is_all_inclusive, :created_at)
		ON CONFLICT (name, location) DO UPDATE SET
			description = EXCLUDED.description,
			link = EXCLUDED.link,
			rating = EXCLUDED.rating,
			is_all_inclusive = EXCLUDED.is_all_inclusive,
			created_at = EXCLUDED.created_at
	`,
	insertHotelRatings: `
		INSERT INTO hotel_ratings (hotel_id, rating, review_count, created_at)
		VALUES (:hotel_id, :rating, :review_count, :created_at)
		ON CONFLICT (hotel_id, rating) DO UPDATE SET
			review_count = EXCLUDED.review_count,
			created_at = EXCLUDED.created_at
	`,
	insertHotelReviews: `
		INSERT INTO hotel_reviews (hotel_id, name, positive, negative, neutral, external_link, created_at)
		VALUES (:hotel_id, :name, :positive, :negative, :neutral, :external_link, :created_at)
		ON CONFLICT (hotel_id, name) DO UPDATE SET
			positive = EXCLUDED.positive,
			negative = EXCLUDED.negative,
			neutral = EXCLUDED.neutral,
			external_link = EXCLUDED.external_link,
			created_at = EXCLUDED.created_at
	`,
	insertPrices: `
		INSERT INTO prices (hotel_id, price, start_date, end_date, currency, checkin_time, checkout_time, property_details_link, created_at)
		VALUES (:hotel_id, :price, :start_date, :end_date, :currency, :checkin_time, :checkout_time, :property_details_link, :created_at)
		ON CONFLICT (hotel_id, start_date, end_date) DO UPDATE SET
			price = EXCLUDED.price,
			currency = EXCLUDED.currency,
			checkin_time = EXCLUDED.checkin_time,
			checkout_time = EXCLUDED.checkout_time,
			property_details_link = EXCLUDED.property_details_link,
			created_at = EXCLUDED.created_at
	`,
	getCurrentPrice: `
		SELECT id, hotel_id, price, start_date, end_date, currency, checkin_time, checkout_time, property_details_link, created_at
		FROM prices
		WHERE hotel_id = $1 AND start_date = $2 AND end_date = $3
		LIMIT 1
	`,
	insertUser: `
		INSERT INTO users (email, password_hash, password_salt, google_id, display_name, avatar_url, created_at)
		VALUES (:email, :password_hash, :password_salt, :google_id, :display_name, :avatar_url, :created_at)
		ON CONFLICT (email) DO UPDATE SET
			password_hash = COALESCE(EXCLUDED.password_hash, users.password_hash),
			password_salt = COALESCE(EXCLUDED.password_salt, users.password_salt),
			google_id = COALESCE(EXCLUDED.google_id, users.google_id),
			display_name = COALESCE(EXCLUDED.display_name, users.display_name),
			avatar_url = COALESCE(EXCLUDED.avatar_url, users.avatar_url),
			created_at = EXCLUDED.created_at
	`,
	insertUserSettings: `
		INSERT INTO user_settings (user_id, setting_key, setting_value, created_at)
		VALUES (:user_id, :setting_key, :setting_value, :created_at)
		ON CONFLICT (user_id, setting_key) DO UPDATE SET
			setting_value = EXCLUDED.setting_value,
			created_at = EXCLUDED.created_at
	`,
	listHotels: `
		SELECT id, name, description, link, rating, location, is_all_inclusive, created_at
		FROM hotels
		ORDER BY name
	`,
	getHotelByID: `
		SELECT id, name, description, link, rating, location, is_all_inclusive, created_at
		FROM hotels
		WHERE id = $1
	`,
	listHotelsByLocation: `
		SELECT id, name, description, link, rating, location, is_all_inclusive, created_at
		FROM hotels
		WHERE location ILIKE '%' || $1 || '%'
		ORDER BY name
	`,
	listHotelRatings: `
		SELECT id, hotel_id, rating, review_count, created_at
		FROM hotel_ratings
		WHERE hotel_id = $1
		ORDER BY created_at DESC
	`,
	listHotelReviews: `
		SELECT id, hotel_id, name, positive, negative, neutral, external_link, created_at
		FROM hotel_reviews
		WHERE hotel_id = $1
		ORDER BY name
	`,
	listPrices: `
		SELECT id, hotel_id, price, start_date, end_date, currency, checkin_time, checkout_time, property_details_link, created_at
		FROM prices
		WHERE hotel_id = $1
		ORDER BY start_date
	`,
	listHotelsPaginated: `
		SELECT id, name, description, link, rating, location, is_all_inclusive, created_at
		FROM hotels
		ORDER BY name
		LIMIT $1 OFFSET $2
	`,
	countHotels: `
		SELECT COUNT(*) FROM hotels
	`,
	listHotelsByLocationPaginated: `
		SELECT id, name, description, link, rating, location, is_all_inclusive, created_at
		FROM hotels
		WHERE location ILIKE '%' || $1 || '%'
		ORDER BY name
		LIMIT $2 OFFSET $3
	`,
	countHotelsByLocation: `
		SELECT COUNT(*) FROM hotels WHERE location ILIKE '%' || $1 || '%'
	`,
	listHotelsWithPrices: `
		SELECT h.id, h.name, h.description, h.link, h.rating, h.location, h.is_all_inclusive, h.created_at,
			COALESCE(MIN(p.price), 0) AS min_price,
			COALESCE(MAX(p.price), 0) AS max_price,
			COALESCE(MIN(p.currency), '') AS price_currency,
			COUNT(p.id) AS price_count
		FROM hotels h
		LEFT JOIN prices p ON p.hotel_id = h.id
		GROUP BY h.id
		ORDER BY h.name
	`,
	listHotelsWithPricesPaginated: `
		SELECT h.id, h.name, h.description, h.link, h.rating, h.location, h.is_all_inclusive, h.created_at,
			COALESCE(MIN(p.price), 0) AS min_price,
			COALESCE(MAX(p.price), 0) AS max_price,
			COALESCE(MIN(p.currency), '') AS price_currency,
			COUNT(p.id) AS price_count
		FROM hotels h
		LEFT JOIN prices p ON p.hotel_id = h.id
		GROUP BY h.id
		ORDER BY h.name
		LIMIT $1 OFFSET $2
	`,
	countHotelsWithPrices: `
		SELECT COUNT(*) FROM (SELECT h.id FROM hotels h LEFT JOIN prices p ON p.hotel_id = h.id GROUP BY h.id) sub
	`,
	listHotelsWithPricesByLocationPaginated: `
		SELECT h.id, h.name, h.description, h.link, h.rating, h.location, h.is_all_inclusive, h.created_at,
			COALESCE(MIN(p.price), 0) AS min_price,
			COALESCE(MAX(p.price), 0) AS max_price,
			COALESCE(MIN(p.currency), '') AS price_currency,
			COUNT(p.id) AS price_count
		FROM hotels h
		LEFT JOIN prices p ON p.hotel_id = h.id
		WHERE h.location ILIKE '%' || $1 || '%'
		GROUP BY h.id
		ORDER BY h.name
		LIMIT $2 OFFSET $3
	`,
	countHotelsWithPricesByLocation: `
		SELECT COUNT(*) FROM (SELECT h.id FROM hotels h LEFT JOIN prices p ON p.hotel_id = h.id WHERE h.location ILIKE '%' || $1 || '%' GROUP BY h.id) sub
	`,
	listUserSettings: `
		SELECT id, user_id, setting_key, setting_value, created_at
		FROM user_settings
		WHERE user_id = $1
		ORDER BY setting_key
	`,
	getUserSetting: `
		SELECT id, user_id, setting_key, setting_value, created_at
		FROM user_settings
		WHERE user_id = $1 AND setting_key = $2
	`,
	deleteUserSetting: `
		DELETE FROM user_settings WHERE user_id = $1 AND setting_key = $2
	`,
	getUserByEmail: `
		SELECT id, email, password_hash, password_salt, google_id, display_name, avatar_url, created_at
		FROM users
		WHERE email = $1
	`,
	getUserByID: `
		SELECT id, email, password_hash, password_salt, google_id, display_name, avatar_url, created_at
		FROM users
		WHERE id = $1
	`,
	getUserByGoogleID: `
		SELECT id, email, password_hash, password_salt, google_id, display_name, avatar_url, created_at
		FROM users
		WHERE google_id = $1
	`,
	updateUser: `
		UPDATE users SET
			email = :email,
			password_hash = :password_hash,
			password_salt = :password_salt,
			google_id = :google_id,
			display_name = :display_name,
			avatar_url = :avatar_url
		WHERE id = :id
	`,
}
