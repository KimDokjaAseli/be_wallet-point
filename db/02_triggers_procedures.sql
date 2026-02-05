-- ==================================================
-- Part 2: Triggers & Stored Procedures
-- ==================================================
USE wallet_point;

-- TRIGGERS
DELIMITER $$

-- Trigger to automatically create a wallet when a new user is inserted
CREATE TRIGGER trg_create_wallet_after_user_insert
AFTER INSERT ON users FOR EACH ROW
BEGIN
    INSERT INTO wallets (user_id, balance, created_at, updated_at)
    VALUES (NEW.id, 0, NOW(), NOW());
END$$

-- Trigger to audit user insertions
CREATE TRIGGER trg_audit_user_insert
AFTER INSERT ON users FOR EACH ROW
BEGIN
    INSERT INTO audit_logs (user_id, action, table_name, record_id, new_value, created_at)
    VALUES (NEW.id, 'INSERT', 'users', NEW.id,
        JSON_OBJECT('email', NEW.email, 'full_name', NEW.full_name, 'nim_nip', NEW.nim_nip, 'role', NEW.role, 'status', NEW.status), NOW());
END$$

-- STORED PROCEDURES
-- Simple procedure to calculate balance from transactions
CREATE PROCEDURE sp_get_wallet_balance(IN p_wallet_id BIGINT UNSIGNED, OUT p_balance INT)
BEGIN
    SELECT COALESCE(SUM(CASE WHEN direction = 'credit' THEN amount WHEN direction = 'debit' THEN -amount END), 0) 
    INTO p_balance FROM wallet_transactions WHERE wallet_id = p_wallet_id AND status = 'success';
END$$

DELIMITER ;
