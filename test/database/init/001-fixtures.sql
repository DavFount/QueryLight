-- Synthetic test data only. Runs once when each database volume is initialized.
USE querylight_test;
SET NAMES utf8mb4;
SET time_zone = '+00:00';

CREATE TABLE value_samples (
    id INT PRIMARY KEY,
    nullable_text VARCHAR(100) NULL,
    empty_text VARCHAR(100) NOT NULL,
    signed_bigint BIGINT NOT NULL,
    unsigned_bigint BIGINT UNSIGNED NOT NULL,
    exact_decimal DECIMAL(30,10) NOT NULL,
    date_value DATE NOT NULL,
    datetime_value DATETIME(6) NOT NULL,
    time_value TIME(6) NOT NULL,
    json_value JSON NOT NULL,
    binary_value VARBINARY(32) NOT NULL,
    bit_value BIT(8) NOT NULL,
    unicode_text VARCHAR(100) NOT NULL
) ENGINE=InnoDB;

INSERT INTO value_samples VALUES
(1, NULL, '', -9223372036854775808, 18446744073709551615,
 12345678901234567890.1234500000, '2026-10-15', '2026-10-15 12:34:56.123456',
 '-25:01:02.123456', '{"large_integer":"18446744073709551615","null":null}',
 X'0001FFFE4142', b'10100101', 'Querylight — café 雪 🎮'),
(2, 'NULL', '', 9223372036854775807, 9007199254740993,
 -0.0000000001, '2000-02-29', '2000-02-29 00:00:00.000001',
 '00:00:00.000000', '[true,false,"",0]', X'', b'00000000', 'Hello RedM');

-- Only this initialization session permits zero dates; server defaults stay intact.
SET @previous_sql_mode = @@SESSION.sql_mode;
SET SESSION sql_mode = '';
CREATE TABLE temporal_edge_cases (id INT PRIMARY KEY, zero_date DATE, zero_datetime DATETIME(6));
INSERT INTO temporal_edge_cases VALUES (1, '0000-00-00', '0000-00-00 00:00:00.000000');
SET SESSION sql_mode = @previous_sql_mode;

CREATE TABLE empty_table (id INT PRIMARY KEY, label VARCHAR(100));
CREATE TABLE `odd``table name` (`odd``column` VARCHAR(100));
INSERT INTO `odd``table name` VALUES ('Identifier quoting fixture');

CREATE TABLE grid_rows (id INT PRIMARY KEY, label VARCHAR(100) NOT NULL);
INSERT INTO grid_rows
SELECT n + 1, CONCAT('Synthetic row ', n + 1)
FROM (
    SELECT ones.n + tens.n * 10 + hundreds.n * 100 AS n
    FROM (SELECT 0 n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4 UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9) ones
    CROSS JOIN (SELECT 0 n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4 UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9) tens
    CROSS JOIN (SELECT 0 n UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4 UNION ALL SELECT 5 UNION ALL SELECT 6 UNION ALL SELECT 7 UNION ALL SELECT 8 UNION ALL SELECT 9) hundreds
) numbers;

CREATE TABLE large_cells (id INT PRIMARY KEY, payload LONGTEXT NOT NULL);
INSERT INTO large_cells VALUES (1, REPEAT('x', 300000));

-- Last table is a sentinel that fixture initialization completed.
CREATE TABLE fixture_ready (id INT PRIMARY KEY, fixture_version INT NOT NULL);
INSERT INTO fixture_ready VALUES (1, 1);
