classDiagram
direction BT
class accounts {
   varchar address
   numeric balance
   bigint nonce
   numeric stake
   timestamp created_at
   integer id
}
class api_keys {
   integer user_id
   varchar key
   jsonb permissions
   timestamp created_at
   timestamp expires_at
   integer id
}
class blocks {
   varchar hash
   varchar prev_hash
   timestamp timestamp
   integer validator_id
   integer tx_count
   varchar state_root
   varchar signature
   integer id
}
class contracts {
   varchar address
   varchar owner
   bytea code
   jsonb abi
   timestamp created_at
   varchar type
   integer id
}
class logs {
   integer tx_id
   timestamp tx_timestamp
   integer contract_id
   varchar event
   jsonb data
   integer id
   timestamp timestamp
}
class logs_2025_06 {
   integer tx_id
   timestamp tx_timestamp
   integer contract_id
   varchar event
   jsonb data
   integer id
   timestamp timestamp
}
class logs_2025_07 {
   integer tx_id
   timestamp tx_timestamp
   integer contract_id
   varchar event
   jsonb data
   integer id
   timestamp timestamp
}
class logs_2025_08 {
   integer tx_id
   timestamp tx_timestamp
   integer contract_id
   varchar event
   jsonb data
   integer id
   timestamp timestamp
}
class logs_2025_09 {
   integer tx_id
   timestamp tx_timestamp
   integer contract_id
   varchar event
   jsonb data
   integer id
   timestamp timestamp
}
class logs_2025_10 {
   integer tx_id
   timestamp tx_timestamp
   integer contract_id
   varchar event
   jsonb data
   integer id
   timestamp timestamp
}
class logs_2025_11 {
   integer tx_id
   timestamp tx_timestamp
   integer contract_id
   varchar event
   jsonb data
   integer id
   timestamp timestamp
}
class logs_2025_12 {
   integer tx_id
   timestamp tx_timestamp
   integer contract_id
   varchar event
   jsonb data
   integer id
   timestamp timestamp
}
class logs_2026_01 {
   integer tx_id
   timestamp tx_timestamp
   integer contract_id
   varchar event
   jsonb data
   integer id
   timestamp timestamp
}
class logs_2026_02 {
   integer tx_id
   timestamp tx_timestamp
   integer contract_id
   varchar event
   jsonb data
   integer id
   timestamp timestamp
}
class logs_2026_03 {
   integer tx_id
   timestamp tx_timestamp
   integer contract_id
   varchar event
   jsonb data
   integer id
   timestamp timestamp
}
class logs_2026_04 {
   integer tx_id
   timestamp tx_timestamp
   integer contract_id
   varchar event
   jsonb data
   integer id
   timestamp timestamp
}
class logs_2026_05 {
   integer tx_id
   timestamp tx_timestamp
   integer contract_id
   varchar event
   jsonb data
   integer id
   timestamp timestamp
}
class metrics {
   varchar metric
   numeric value
   timestamp timestamp
   integer id
}
class oracles {
   varchar name
   varchar url
   varchar status
   timestamp last_check
   integer id
}
class poa_validators {
   integer validator_id
   varchar legal_name
   varchar registration_number
   varchar jurisdiction
   jsonb contact_info
   timestamp approval_date
   jsonb poa_metadata
   integer id
}
class pos_validators {
   integer validator_id
   numeric total_stake
   numeric delegated_stake
   numeric commission_rate
   numeric uptime
   integer slashing_events
   numeric rewards
   timestamp last_reward_at
   jsonb pos_metadata
   integer id
}
class token_balances {
   integer token_id
   varchar address
   numeric balance
   integer id
}
class tokens {
   integer contract_id
   varchar standard
   varchar symbol
   varchar name
   integer decimals
   numeric total_supply
   integer id
}
class transactions {
   integer block_id
   varchar hash
   varchar sender
   varchar recipient
   numeric value
   numeric fee
   bigint nonce
   varchar type
   integer contract_id
   jsonb payload
   varchar status
   integer id
   timestamp timestamp
}
class transactions_2025_06 {
   integer block_id
   varchar hash
   varchar sender
   varchar recipient
   numeric value
   numeric fee
   bigint nonce
   varchar type
   integer contract_id
   jsonb payload
   varchar status
   integer id
   timestamp timestamp
}
class transactions_2025_07 {
   integer block_id
   varchar hash
   varchar sender
   varchar recipient
   numeric value
   numeric fee
   bigint nonce
   varchar type
   integer contract_id
   jsonb payload
   varchar status
   integer id
   timestamp timestamp
}
class transactions_2025_08 {
   integer block_id
   varchar hash
   varchar sender
   varchar recipient
   numeric value
   numeric fee
   bigint nonce
   varchar type
   integer contract_id
   jsonb payload
   varchar status
   integer id
   timestamp timestamp
}
class transactions_2025_09 {
   integer block_id
   varchar hash
   varchar sender
   varchar recipient
   numeric value
   numeric fee
   bigint nonce
   varchar type
   integer contract_id
   jsonb payload
   varchar status
   integer id
   timestamp timestamp
}
class transactions_2025_10 {
   integer block_id
   varchar hash
   varchar sender
   varchar recipient
   numeric value
   numeric fee
   bigint nonce
   varchar type
   integer contract_id
   jsonb payload
   varchar status
   integer id
   timestamp timestamp
}
class transactions_2025_11 {
   integer block_id
   varchar hash
   varchar sender
   varchar recipient
   numeric value
   numeric fee
   bigint nonce
   varchar type
   integer contract_id
   jsonb payload
   varchar status
   integer id
   timestamp timestamp
}
class transactions_2025_12 {
   integer block_id
   varchar hash
   varchar sender
   varchar recipient
   numeric value
   numeric fee
   bigint nonce
   varchar type
   integer contract_id
   jsonb payload
   varchar status
   integer id
   timestamp timestamp
}
class transactions_2026_01 {
   integer block_id
   varchar hash
   varchar sender
   varchar recipient
   numeric value
   numeric fee
   bigint nonce
   varchar type
   integer contract_id
   jsonb payload
   varchar status
   integer id
   timestamp timestamp
}
class transactions_2026_02 {
   integer block_id
   varchar hash
   varchar sender
   varchar recipient
   numeric value
   numeric fee
   bigint nonce
   varchar type
   integer contract_id
   jsonb payload
   varchar status
   integer id
   timestamp timestamp
}
class transactions_2026_03 {
   integer block_id
   varchar hash
   varchar sender
   varchar recipient
   numeric value
   numeric fee
   bigint nonce
   varchar type
   integer contract_id
   jsonb payload
   varchar status
   integer id
   timestamp timestamp
}
class transactions_2026_04 {
   integer block_id
   varchar hash
   varchar sender
   varchar recipient
   numeric value
   numeric fee
   bigint nonce
   varchar type
   integer contract_id
   jsonb payload
   varchar status
   integer id
   timestamp timestamp
}
class transactions_2026_05 {
   integer block_id
   varchar hash
   varchar sender
   varchar recipient
   numeric value
   numeric fee
   bigint nonce
   varchar type
   integer contract_id
   jsonb payload
   varchar status
   integer id
   timestamp timestamp
}
class validators {
   varchar address
   varchar pubkey
   numeric stake
   varchar status
   timestamp last_active
   varchar consensus_type
   jsonb authority_info
   jsonb metadata
   integer id
}
class wallets {
   integer account_id
   varchar public_key
   varchar private_key
   timestamp created_at
   timestamp updated_at
   varchar status
   integer id
}

blocks  -->  validators : validator_id:id
logs  -->  contracts : contract_id:id
logs  -->  transactions : tx_id, tx_timestamp:id, timestamp
logs  -->  transactions_2025_06 : tx_id, tx_timestamp:id, timestamp
logs  -->  transactions_2025_07 : tx_id, tx_timestamp:id, timestamp
logs  -->  transactions_2025_08 : tx_id, tx_timestamp:id, timestamp
logs  -->  transactions_2025_09 : tx_id, tx_timestamp:id, timestamp
logs  -->  transactions_2025_10 : tx_id, tx_timestamp:id, timestamp
logs  -->  transactions_2025_11 : tx_id, tx_timestamp:id, timestamp
logs  -->  transactions_2025_12 : tx_id, tx_timestamp:id, timestamp
logs  -->  transactions_2026_01 : tx_id, tx_timestamp:id, timestamp
logs  -->  transactions_2026_02 : tx_id, tx_timestamp:id, timestamp
logs  -->  transactions_2026_03 : tx_id, tx_timestamp:id, timestamp
logs  -->  transactions_2026_04 : tx_id, tx_timestamp:id, timestamp
logs  -->  transactions_2026_05 : tx_id, tx_timestamp:id, timestamp
logs_2025_06  -->  contracts : contract_id:id
logs_2025_06  -->  transactions : tx_id, tx_timestamp:id, timestamp
logs_2025_07  -->  contracts : contract_id:id
logs_2025_07  -->  transactions : tx_id, tx_timestamp:id, timestamp
logs_2025_08  -->  contracts : contract_id:id
logs_2025_08  -->  transactions : tx_id, tx_timestamp:id, timestamp
logs_2025_09  -->  contracts : contract_id:id
logs_2025_09  -->  transactions : tx_id, tx_timestamp:id, timestamp
logs_2025_10  -->  contracts : contract_id:id
logs_2025_10  -->  transactions : tx_id, tx_timestamp:id, timestamp
logs_2025_11  -->  contracts : contract_id:id
logs_2025_11  -->  transactions : tx_id, tx_timestamp:id, timestamp
logs_2025_12  -->  contracts : contract_id:id
logs_2025_12  -->  transactions : tx_id, tx_timestamp:id, timestamp
logs_2026_01  -->  contracts : contract_id:id
logs_2026_01  -->  transactions : tx_id, tx_timestamp:id, timestamp
logs_2026_02  -->  contracts : contract_id:id
logs_2026_02  -->  transactions : tx_id, tx_timestamp:id, timestamp
logs_2026_03  -->  contracts : contract_id:id
logs_2026_03  -->  transactions : tx_id, tx_timestamp:id, timestamp
logs_2026_04  -->  contracts : contract_id:id
logs_2026_04  -->  transactions : tx_id, tx_timestamp:id, timestamp
logs_2026_05  -->  contracts : contract_id:id
logs_2026_05  -->  transactions : tx_id, tx_timestamp:id, timestamp
poa_validators  -->  validators : validator_id:id
pos_validators  -->  validators : validator_id:id
token_balances  -->  tokens : token_id:id
tokens  -->  contracts : contract_id:id
transactions  -->  blocks : block_id:id
transactions  -->  contracts : contract_id:id
transactions_2025_06  -->  blocks : block_id:id
transactions_2025_06  -->  contracts : contract_id:id
transactions_2025_07  -->  blocks : block_id:id
transactions_2025_07  -->  contracts : contract_id:id
transactions_2025_08  -->  blocks : block_id:id
transactions_2025_08  -->  contracts : contract_id:id
transactions_2025_09  -->  blocks : block_id:id
transactions_2025_09  -->  contracts : contract_id:id
transactions_2025_10  -->  blocks : block_id:id
transactions_2025_10  -->  contracts : contract_id:id
transactions_2025_11  -->  blocks : block_id:id
transactions_2025_11  -->  contracts : contract_id:id
transactions_2025_12  -->  blocks : block_id:id
transactions_2025_12  -->  contracts : contract_id:id
transactions_2026_01  -->  blocks : block_id:id
transactions_2026_01  -->  contracts : contract_id:id
transactions_2026_02  -->  blocks : block_id:id
transactions_2026_02  -->  contracts : contract_id:id
transactions_2026_03  -->  blocks : block_id:id
transactions_2026_03  -->  contracts : contract_id:id
transactions_2026_04  -->  blocks : block_id:id
transactions_2026_04  -->  contracts : contract_id:id
transactions_2026_05  -->  blocks : block_id:id
transactions_2026_05  -->  contracts : contract_id:id
wallets  -->  accounts : account_id:id
