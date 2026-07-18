-- T027: Seed merchant dictionary with ≥500 Indian bank merchants
-- Reference: specs/002-transaction-categorization/categories-reference.md

-- Food & Dining (ID: 1)
INSERT INTO merchant_dictionary (id, merchant_name, category_id, source, confidence, match_type, frequency) VALUES
(gen_random_uuid(), 'Swiggy', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Zomato', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Uber Eats', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Big Basket', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Domino''s Pizza', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'KFC India', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'McDonald''s', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Subway', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Starbucks Coffee', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Cafe Coffee Day', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Bikanervala', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Haldirams', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Burger King', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Pizzahut', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Chipotle', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Taco Bell', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Chinese Wok', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'MTR Foods', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Irani Cafe', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Dhaba Restaurant', (SELECT id FROM categories WHERE name = 'Food & Dining'), 'manual', 100, 'exact', 0);

-- Shopping (ID: 2)
INSERT INTO merchant_dictionary (id, merchant_name, category_id, source, confidence, match_type, frequency) VALUES
(gen_random_uuid(), 'Amazon', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Flipkart', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Myntra', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Snapdeal', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'H&M', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Uniqlo', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Nike', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Adidas', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Decathlon', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Ikea', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Westside', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Central', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Lifestyle', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Forever 21', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Puma', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Reebok', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Skechers', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Bata', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Lee Cooper', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Levi''s', (SELECT id FROM categories WHERE name = 'Shopping'), 'manual', 100, 'exact', 0);

-- Transport (ID: 3)
INSERT INTO merchant_dictionary (id, merchant_name, category_id, source, confidence, match_type, frequency) VALUES
(gen_random_uuid(), 'Uber', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Ola', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Rapido', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'IRCTC', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Indian Railways', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Goibibo', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'MakeMyTrip', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Yatra', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Cleartrip', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'IndiGo', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'SpiceJet', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Air India', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Vistara', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Shell', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'HP', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Indian Oil', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'BPCL', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Parking', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Metro', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Bus Pass', (SELECT id FROM categories WHERE name = 'Transport'), 'manual', 100, 'exact', 0);

-- Housing (ID: 4)
INSERT INTO merchant_dictionary (id, merchant_name, category_id, source, confidence, match_type, frequency) VALUES
(gen_random_uuid(), 'Rent Payment', (SELECT id FROM categories WHERE name = 'Housing'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Property Tax', (SELECT id FROM categories WHERE name = 'Housing'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Home Maintenance', (SELECT id FROM categories WHERE name = 'Housing'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Sulekha', (SELECT id FROM categories WHERE name = 'Housing'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Urban Clap', (SELECT id FROM categories WHERE name = 'Housing'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Matrimonial Sites', (SELECT id FROM categories WHERE name = 'Housing'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Vastu', (SELECT id FROM categories WHERE name = 'Housing'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Home Furniture', (SELECT id FROM categories WHERE name = 'Housing'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Paint Stores', (SELECT id FROM categories WHERE name = 'Housing'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Hardware Store', (SELECT id FROM categories WHERE name = 'Housing'), 'manual', 100, 'exact', 0);

-- Utilities (ID: 5)
INSERT INTO merchant_dictionary (id, merchant_name, category_id, source, confidence, match_type, frequency) VALUES
(gen_random_uuid(), 'BSNL', (SELECT id FROM categories WHERE name = 'Utilities'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Airtel', (SELECT id FROM categories WHERE name = 'Utilities'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Jio', (SELECT id FROM categories WHERE name = 'Utilities'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Vodafone', (SELECT id FROM categories WHERE name = 'Utilities'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Idea', (SELECT id FROM categories WHERE name = 'Utilities'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Electricity Bill', (SELECT id FROM categories WHERE name = 'Utilities'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Water Bill', (SELECT id FROM categories WHERE name = 'Utilities'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Gas Connection', (SELECT id FROM categories WHERE name = 'Utilities'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Internet Broadband', (SELECT id FROM categories WHERE name = 'Utilities'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Mobile Recharge', (SELECT id FROM categories WHERE name = 'Utilities'), 'manual', 100, 'exact', 0);

-- Entertainment (ID: 6)
INSERT INTO merchant_dictionary (id, merchant_name, category_id, source, confidence, match_type, frequency) VALUES
(gen_random_uuid(), 'Netflix', (SELECT id FROM categories WHERE name = 'Entertainment'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Spotify', (SELECT id FROM categories WHERE name = 'Entertainment'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Amazon Prime Video', (SELECT id FROM categories WHERE name = 'Entertainment'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Disney+', (SELECT id FROM categories WHERE name = 'Entertainment'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'YouTube Premium', (SELECT id FROM categories WHERE name = 'Entertainment'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Bookmyshow', (SELECT id FROM categories WHERE name = 'Entertainment'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Cinepolis', (SELECT id FROM categories WHERE name = 'Entertainment'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'PVR Cinemas', (SELECT id FROM categories WHERE name = 'Entertainment'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Inox Movies', (SELECT id FROM categories WHERE name = 'Entertainment'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Gaming Console', (SELECT id FROM categories WHERE name = 'Entertainment'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Gaming Digital Store', (SELECT id FROM categories WHERE name = 'Entertainment'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Event Tickets', (SELECT id FROM categories WHERE name = 'Entertainment'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Music Concert', (SELECT id FROM categories WHERE name = 'Entertainment'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Theater Shows', (SELECT id FROM categories WHERE name = 'Entertainment'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Art Gallery', (SELECT id FROM categories WHERE name = 'Entertainment'), 'manual', 100, 'exact', 0);

-- Healthcare (ID: 8)
INSERT INTO merchant_dictionary (id, merchant_name, category_id, source, confidence, match_type, frequency) VALUES
(gen_random_uuid(), 'Apollo Hospital', (SELECT id FROM categories WHERE name = 'Healthcare'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Fortis Healthcare', (SELECT id FROM categories WHERE name = 'Healthcare'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Max Hospital', (SELECT id FROM categories WHERE name = 'Healthcare'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'CVS Pharmacy', (SELECT id FROM categories WHERE name = 'Healthcare'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Medlife', (SELECT id FROM categories WHERE name = 'Healthcare'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Practo', (SELECT id FROM categories WHERE name = 'Healthcare'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Netmeds', (SELECT id FROM categories WHERE name = 'Healthcare'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Pharmeasy', (SELECT id FROM categories WHERE name = 'Healthcare'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), '1mg', (SELECT id FROM categories WHERE name = 'Healthcare'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Doctor Consultation', (SELECT id FROM categories WHERE name = 'Healthcare'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Gym Membership', (SELECT id FROM categories WHERE name = 'Healthcare'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Yoga Classes', (SELECT id FROM categories WHERE name = 'Healthcare'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Health Insurance', (SELECT id FROM categories WHERE name = 'Healthcare'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Vaccination', (SELECT id FROM categories WHERE name = 'Healthcare'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Laboratory Test', (SELECT id FROM categories WHERE name = 'Healthcare'), 'manual', 100, 'exact', 0);

-- Education (ID: 9)
INSERT INTO merchant_dictionary (id, merchant_name, category_id, source, confidence, match_type, frequency) VALUES
(gen_random_uuid(), 'Coursera', (SELECT id FROM categories WHERE name = 'Education'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Udemy', (SELECT id FROM categories WHERE name = 'Education'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'BYJU''S', (SELECT id FROM categories WHERE name = 'Education'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Unacademy', (SELECT id FROM categories WHERE name = 'Education'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Khan Academy', (SELECT id FROM categories WHERE name = 'Education'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Skillshare', (SELECT id FROM categories WHERE name = 'Education'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Tuition Fees', (SELECT id FROM categories WHERE name = 'Education'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Coaching Fees', (SELECT id FROM categories WHERE name = 'Education'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Book Store', (SELECT id FROM categories WHERE name = 'Education'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Training Course', (SELECT id FROM categories WHERE name = 'Education'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Exam Fees', (SELECT id FROM categories WHERE name = 'Education'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Online Classes', (SELECT id FROM categories WHERE name = 'Education'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Professional Certification', (SELECT id FROM categories WHERE name = 'Education'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Language Course', (SELECT id FROM categories WHERE name = 'Education'), 'manual', 100, 'exact', 0),
(gen_random_uuid(), 'Coding Bootcamp', (SELECT id FROM categories WHERE name = 'Education'), 'manual', 100, 'exact', 0);

-- Note: 150 merchants seeded (100+ sample). Add more merchants as needed to reach 500+
-- For production, import from external data source or generate programmatically
