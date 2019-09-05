CREATE DATABASE ledger_bridge_cache WITH OWNER=mosoly LC_COLLATE ='en_US.utf8' LC_CTYPE ='en_US.utf8' ENCODING ='UTF8';

CREATE TABLE IF NOT EXISTS transaction_states
(
    id   INTEGER NOT NULL
        CONSTRAINT transaction_states_id_pk
            PRIMARY KEY,
    status TEXT    NOT NULL
);

INSERT INTO transaction_states (id, status) VALUES (1, 'IN_PROGRESS');
INSERT INTO transaction_states (id, status) VALUES (2, 'SUCCESS');
INSERT INTO transaction_states (id, status) VALUES (3, 'FAILED');

CREATE TABLE IF NOT EXISTS transactions
(
    id                  BIGSERIAL NOT NULL
        CONSTRAINT transactions_id_pk
            PRIMARY KEY,
    transaction_hash      TEXT NOT NULL ,
    transaction_state_id  INTEGER NOT NULL
        CONSTRAINT transactions_transaction_state_id
            REFERENCES transaction_states,
    created               TIMESTAMP NOT NULL,
    updated               TIMESTAMP NOT NULL,
    modified_by           TEXT
);

CREATE TABLE IF NOT EXISTS ethereum_blockchain
(
    id BIGINT NOT NULL
        CONSTRAINT ethereum_blockchain_pk
            PRIMARY KEY,
    latest_processed_block_number BIGINT
);

INSERT INTO ethereum_blockchain(id, latest_processed_block_number) VALUES (1, 6313390); -- block number mined as of Sep-02-2019 01:14:46 PM +UTC

CREATE TABLE IF NOT EXISTS user_data
(
    id BIGINT NOT NULL
        CONSTRAINT user_data_id_pk PRIMARY KEY,
    account TEXT NOT NULL
        CONSTRAINT user_data_account UNIQUE,
    transaction_id BIGINT NULL
        CONSTRAINT user_data_transaction_id
            REFERENCES transactions,
    invite_url_hash TEXT NOT NULL,
    validated BOOLEAN NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS mentorship
(
    user_id BIGINT NOT NULL
        CONSTRAINT mentorship_user_id
            REFERENCES user_data,
    mentoree_id BIGINT NOT NULL
        CONSTRAINT mentorship_mentoree_id
            REFERENCES user_data,
    transaction_id BIGINT NULL
        CONSTRAINT mentorship_transaction_id
            REFERENCES transactions,
    CONSTRAINT mentorship_user_id_mentoree_id_unique UNIQUE (user_id, mentoree_id)
);

-- Partially implemented:

CREATE TABLE IF NOT EXISTS project_data
(
    id BIGINT NOT NULL
        CONSTRAINT project_data_id_pk PRIMARY KEY,
    transaction_id BIGINT NULL
        CONSTRAINT project_data_transaction_id
            REFERENCES transactions,
    name TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    passport_address TEXT NULL
        CONSTRAINT project_data_passport_address UNIQUE
);
