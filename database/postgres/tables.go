package postgres

const (
	tableNameAddresses           = `"Addresses"`
	tableNameBlocks              = `"Blocks"`
	tableNameUncles              = `"Uncles"`
	tableNameBytecodes           = `"Bytecodes"`
	tableNameTransactions        = `"Transactions"`
	tableNameStorageKeys         = `"StorageKeys"`
	tableNameContracts           = `"Contracts"`
	tableNameReceipts            = `"Receipts"`
	tableNameEventTypes          = `"FirstTopics"`
	tableNameLogs                = `"Logs"`
	tableNameErc20Tokens         = `"Erc20Tokens"`
	tableNameErc20TokenTransfers = `"Erc20TokenTransfers"`
	tableNameTraces              = `"Traces"`
	tableNameEtherBalances       = `"EtherBalances"`
	tableNameErc20TokenBalances  = `"Erc20TokenBalances"`
	tableNameStateChanges        = `"StateChanges"`
	tableNammeStorageChanges     = `"StorageChanges"`
)

var tableColumnsBlocks = []string{
	"hash",
	"number",
	"nonce",
	"sha3_uncles",
	"logs_bloom",
	"state_root",
	"miner",
	"difficulty",
	"total_difficulty",
	"size",
	"extra_data",
	"gas_limit",
	"gas_used",
	"base_fee",
	"mix_hash",
	"static_reward",
	"timestamp"}

var tableColumnsUncles = []string{
	"hash",
	"position",
	"uncle_height",
	"block_height",
	"parent_hash",
	// "sha3_uncles",
	"miner",
	"difficulty",
	"gas_limit",
	"gas_used",
	"reward",
	"timestamp"}

var tableColumnsContracts = []string{
	"address_id", // query depends on this being at index 0 in this slice
	"transaction_id",
	"bytecode_id"}

var tableColumnsTransactions = []string{
	"id", // query depends on this being at index 0 in this slice
	"block_id",
	"hash",
	"nonce",
	"index",
	"from_address_id",
	"to_address_id",
	"value",
	"gas",
	"gas_price",
	"gas_tip_cap",
	"gas_fee_cap",
	"input",
	"transaction_type"}

var tableColumnsStorageKeys = []string{
	"transaction_id", // query depends on this being at index 0 in this slice
	"address_id",
	"storage_key"}

var tableColumnsErc20Tokens = []string{
	"address_id", // query depends on this being at index 0 in this slice
	"symbol",
	"name",
	"decimals",
	"total_supply"}

var tableColumnsBytecodes = []string{
	"bytecode",
	"sha256"}

var tableColumnsLogs = []string{
	"transaction_id",
	"index",
	"address_id",
	"topic_0_id",
	"topic_1",
	"topic_2",
	"topic_3",
	"data"}

var tableColumnsReceipts = []string{
	"transaction_id", // query depends on this being at index 0 in this slice
	"cumulative_gas_used",
	"gas_used",
	"contract_address_id",
	"post_state",
	"success",
	"effective_gas_price"}

var tableColumnsEventTypes = []string{
	"id", // query depends on this being at index 0 in this slice
	"hash",
	"signature"}

var tableColumnsErc20TokenTransfers = []string{
	"transaction_id",
	"log_index",
	"token_address_id",
	"from_address_id",
	"to_address_id",
	"value"}

var tableColumnsTraces = []string{
	"transaction_id",
	"index",
	"type",
	"input",
	"from_address_id",
	"to_address_id",
	"value",
	"gas",
	"error"}

var tableColumnsEtherBalances = []string{
	"address_id",
	"block_id",
	"balance"}
var tableColumnsErc20TokenBalances = []string{
	"address_id",
	"block_id",
	"token_address_id",
	"balance"}

var tableColumnsStateChanges = []string{
	"transaction_id", // query depends on this being at index 0 in this slice
	"address_id",
	"balance_before",
	"balance_after",
	"nonce_before",
	"nonce_after"}

var tableColumnsStorageChanges = []string{
	"transaction_id", // query depends on this being at index 0 in this slice
	"address_id",
	"storage_address",
	"value_before",
	"value_after"}
