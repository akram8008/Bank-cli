CREATE TABLE IF NOT EXISTS accounts
(
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    name   TEXT NOT NULL,
    number INTEGER NOT NULL,
    money INTEGER NOT NULL,
    client_id      INTEGER ,
    service_id      INTEGER

);


INSERT INTO accounts (name,number,money,client_id,service_id)  VALUES (?, ?, ?, ?,?);


INSERT INTO accounts (name,number,money,client_id,service_id) VALUES (?, ?, ?, ?, ?);

UPDATE accounts SET number=:a WHERE client_id=:b and service_id=:b;

INSERT INTO accounts VALUES (:id, :name, :number, :money, :client_id) ON CONFLICT DO NOTHING;