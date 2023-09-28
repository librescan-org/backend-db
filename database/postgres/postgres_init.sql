CREATE TABLE IF NOT EXISTS "Addresses"(
    "id" bigserial PRIMARY KEY,
    "hash" bytea NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS "Blocks" (
    "number" bigint PRIMARY KEY,
    "hash" bytea UNIQUE NOT NULL,
    "nonce" bytea NOT NULL,
    "sha3_uncles" bytea NOT NULL,
    "logs_bloom" bytea NOT NULL,
    "state_root" bytea NOT NULL,
    "miner" bigint REFERENCES "Addresses" NOT NULL,
    "difficulty" bytea NOT NULL,
    "total_difficulty" bytea NOT NULL,
    "size" bytea NULL,
    "extra_data" bytea NULL,
    "gas_limit" bytea NULL,
    "gas_used" bytea NULL,
    "base_fee" bytea NULL,
    "mix_hash" bytea NULL,
    "static_reward" bytea NOT NULL,
    "timestamp" bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS "Uncles" (
    "hash" bytea PRIMARY KEY,
    "position" smallint NOT NULL,
    "uncle_height" bigint NOT NULL,
    "block_height" bigint REFERENCES "Blocks" ON DELETE CASCADE NOT NULL,
    "parent_hash" bytea NOT NULL,
    "miner" bigint REFERENCES "Addresses",
    "difficulty" bytea NOT NULL,
    "gas_limit" bigint NULL,
    "gas_used" bigint NULL,
    "reward" bytea NOT NULL,
    "timestamp" bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS "Bytecodes" (
    "id" bigserial PRIMARY KEY,
    "bytecode" bytea NOT NULL,
    "sha256" bytea UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS "Transactions" (
    "id" bigserial PRIMARY KEY,
    "block_id" bigint REFERENCES "Blocks" ON DELETE CASCADE NOT NULL,
    "hash" bytea UNIQUE NOT NULL,
    "nonce" bytea NOT NULL,
    "index" bytea NOT NULL,
    "from_address_id" bigint REFERENCES "Addresses" NOT NULL,
    "to_address_id" bigint REFERENCES "Addresses" NULL,
    "value" bytea NOT NULL,
    "gas" bytea NOT NULL,
    "gas_price" bytea NOT NULL,
    "gas_tip_cap" bytea NULL,
    "gas_fee_cap" bytea NULL,
    "input" bytea NOT NULL,
    "transaction_type" bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS "StorageKeys" (
    "transaction_id" bigint REFERENCES "Transactions" ON DELETE CASCADE NOT NULL,
    "address_id" bigint REFERENCES "Addresses" NOT NULL,
    "storage_key" bytea NOT NULL
);

CREATE TABLE IF NOT EXISTS "Contracts" (
    "address_id" bigint PRIMARY KEY REFERENCES "Addresses",
    "transaction_id" bigint REFERENCES "Transactions" ON DELETE CASCADE NOT NULL,
    "bytecode_id" bigint REFERENCES "Bytecodes" NOT NULL
);

CREATE TABLE IF NOT EXISTS "Receipts" (
    "transaction_id" bigint PRIMARY KEY REFERENCES "Transactions" ON DELETE CASCADE,
    "cumulative_gas_used" bytea NOT NULL,
    "gas_used" bytea NOT NULL,
    "contract_address_id" bigint REFERENCES "Addresses" NULL,
    "post_state" bytea NOT NULL,
    "success" bool NOT NULL,
    "effective_gas_price" bytea NOT NULL
);

CREATE TABLE IF NOT EXISTS "FirstTopics" (
    "id" bigserial PRIMARY KEY,
    "hash" bytea UNIQUE NOT NULL,
    "signature" text NULL
);

CREATE TABLE IF NOT EXISTS "Logs" (
    "transaction_id" bigint REFERENCES "Transactions" ON DELETE CASCADE NOT NULL,
    "index" bytea NOT NULL,
    "address_id" bigint REFERENCES "Addresses" NOT NULL,
    "topic_0_id" bigint REFERENCES "FirstTopics" NULL,
    "topic_1" bytea NOT NULL,
    "topic_2" bytea NOT NULL,
    "topic_3" bytea NOT NULL,
    "data" bytea NOT NULL,
    PRIMARY KEY("transaction_id", "index")
);

CREATE TABLE IF NOT EXISTS "Erc20Tokens" (
    "address_id" bigint PRIMARY KEY REFERENCES "Addresses",
    "symbol" text NULL,
    "name" text NULL,
    "decimals" smallint NULL,
    "total_supply" bytea NULL
);

CREATE TABLE IF NOT EXISTS "Erc20TokenTransfers" (
    "transaction_id" bigint REFERENCES "Transactions" ON DELETE CASCADE NOT NULL,
    "log_index" bytea NOT NULL,
    "token_address_id" bigint REFERENCES "Addresses" NOT NULL,
    "from_address_id" bigint REFERENCES "Addresses" NOT NULL,
    "to_address_id" bigint REFERENCES "Addresses" NOT NULL,
    "value" bytea NOT NULL,
    PRIMARY KEY("transaction_id", "log_index")
);

CREATE TABLE IF NOT EXISTS "Traces" (
    "transaction_id" bigint REFERENCES "Transactions" ON DELETE CASCADE NOT NULL,
    "index" integer NOT NULL,
    "type" text NOT NULL,
    "input" bytea NULL,
    "from_address_id" bigint REFERENCES "Addresses" NOT NULL,
    "to_address_id" bigint REFERENCES "Addresses" NOT NULL,
    "value" bytea NOT NULL,
    "gas" bigint NULL,
    "error" text NULL,
    PRIMARY KEY("transaction_id", "index")
);

CREATE TABLE IF NOT EXISTS "EtherBalances" (
    "address_id" bigint REFERENCES "Addresses" NOT NULL,
    "block_id" bigint REFERENCES "Blocks" ON DELETE CASCADE NOT NULL,
    "balance" bytea NOT NULL,
    PRIMARY KEY("address_id", "block_id")
);

CREATE TABLE IF NOT EXISTS "Erc20TokenBalances" (
    "address_id" bigint REFERENCES "Addresses" NOT NULL,
    "block_id" bigint REFERENCES "Blocks" ON DELETE CASCADE NOT NULL,
    "token_address_id" bigint REFERENCES "Addresses" NOT NULL,
    "balance" bytea NOT NULL,
    PRIMARY KEY("address_id", "block_id", "token_address_id")
);

CREATE TABLE IF NOT EXISTS "StateChanges" (
    "transaction_id" bigint REFERENCES "Transactions" ON DELETE CASCADE NOT NULL,
    "address_id" bigint REFERENCES "Addresses" NOT NULL,
    "balance_before" bytea NOT NULL,
    "balance_after" bytea NOT NULL,
    "nonce_before" bytea NULL,
    "nonce_after" bytea NULL,
    PRIMARY KEY("transaction_id", "address_id")
);

CREATE TABLE IF NOT EXISTS "StorageChanges" (
    "transaction_id" bigint REFERENCES "Transactions" ON DELETE CASCADE NOT NULL,
    "address_id" bigint REFERENCES "Addresses" NOT NULL,
    "storage_address" bytea NOT NULL,
    "value_before" bytea NOT NULL,
    "value_after" bytea NOT NULL,
    PRIMARY KEY("transaction_id", "address_id", "storage_address")
);