-- ==================================================
-- Part 3: Seed Data
-- ==================================================
USE wallet_point;

-- Sample users (password: Password123!)
-- Password hash for 'Password123!' using bcrypt
INSERT INTO users (email, password_hash, full_name, nim_nip, role, status) VALUES
('admin@campus.edu', '$2a$10$D1pwwOkIxUD4zjcgvPpDUOd9uTePy9t/OIkFmTJJRa98TtZw/n152', 'System Administrator', 'ADM001', 'admin', 'active'),
('dosen1@campus.edu', '$2a$10$D1pwwOkIxUD4zjcgvPpDUOd9uTePy9t/OIkFmTJJRa98TtZw/n152', 'Dr. John Doe', 'NIP001', 'dosen', 'active'),
('dosen2@campus.edu', '$2a$10$D1pwwOkIxUD4zjcgvPpDUOd9uTePy9t/OIkFmTJJRa98TtZw/n152', 'Dr. Jane Smith', 'NIP002', 'dosen', 'active'),
('mahasiswa1@campus.edu', '$2a$10$D1pwwOkIxUD4zjcgvPpDUOd9uTePy9t/OIkFmTJJRa98TtZw/n152', 'Alice Johnson', '2023001', 'mahasiswa', 'active'),
('mahasiswa2@campus.edu', '$2a$10$D1pwwOkIxUD4zjcgvPpDUOd9uTePy9t/OIkFmTJJRa98TtZw/n152', 'Bob Williams', '2023002', 'mahasiswa', 'active'),
('mahasiswa3@campus.edu', '$2a$10$D1pwwOkIxUD4zjcgvPpDUOd9uTePy9t/OIkFmTJJRa98TtZw/n152', 'Charlie Brown', '2023003', 'mahasiswa', 'active');

-- Sample products
INSERT INTO products (name, description, price, stock, status, created_by) VALUES
('Notebook', 'Campus branded notebook', 50, 100, 'active', 1),
('Pen Set', 'Premium pen set (5 pcs)', 30, 200, 'active', 1),
('T-Shirt', 'Campus t-shirt (various sizes)', 100, 50, 'active', 1),
('Coffee Voucher', 'Campus cafeteria coffee voucher', 20, 500, 'active', 1);

-- Sample mission questions for a quiz
INSERT INTO missions (creator_id, title, description, type, points_reward, status) VALUES
(2, 'Kuis Pengetahuan Kampus', 'Uji pengetahuanmu tentang sejarah kampus kita.', 'quiz', 100, 'active');

INSERT INTO mission_questions (mission_id, question, options, answer) VALUES
(1, 'Siapa pendiri kampus kita?', '["Bapak A", "Bapak B", "Bapak C", "Bapak D"]', 'Bapak A'),
(1, 'Tahun berapa kampus ini berdiri?', '["1960", "1970", "1980", "1990"]', '1960');
