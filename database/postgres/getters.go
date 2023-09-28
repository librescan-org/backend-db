package postgres

import (
	"database/sql"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	storage "github.com/librescan-org/backend-db"
	// _ "github.com/jackc/pgx/v5/stdlib"
)

func (repo *PostgresRepository) GetEventTypeById(eventTypeId storage.Topic0Id) (*storage.EventType, error) {
	var eventType storage.EventType
	err := repo.statementBuilder.
		Select(tableColumnsEventTypes[1:]...).
		From(tableNameEventTypes).
		Where("id = ?", eventTypeId).
		Scan(&eventType.Hash, &eventType.Signature)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &eventType, err
}
func (repo *PostgresRepository) getBlock(keyColumnName string, key any) (*storage.Block, error) {
	query, _ := repo.statementBuilder.
		Select(tableColumnsBlocks...).
		From(tableNameBlocks).
		Where(fmt.Sprintf(`"%s" = ?`, keyColumnName), key).MustSql()
	return scanBlock(repo.tx.QueryRow(query, key))
}
func (repo *PostgresRepository) GetBlockByHash(hash *common.Hash) (*storage.Block, error) {
	return repo.getBlock("hash", hash)
}
func (repo *PostgresRepository) GetBlockByNumber(number uint64) (*storage.Block, error) {
	return repo.getBlock("number", number)
}
func (repo *PostgresRepository) GetTransactionById(transactionId storage.TransactionId) (*storage.Transaction, error) {
	query, args := repo.statementBuilder.
		Select(tableColumnsTransactions...).
		From(tableNameTransactions).
		Where("id = ?", transactionId).
		MustSql()
	transaction, _, err := scanTransaction(repo.tx.QueryRow(query, args...))
	return transaction, err
}
func (repo *PostgresRepository) GetTransactionByHash(transactionHash *common.Hash) (*storage.Transaction, storage.TransactionId, error) {
	query, args := repo.statementBuilder.
		Select(tableColumnsTransactions...).
		From(tableNameTransactions).
		Where("hash = ?", transactionHash).
		MustSql()
	return scanTransaction(repo.tx.QueryRow(query, args...))
}
func (repo *PostgresRepository) GetAddressById(addressId storage.AddressId) (*common.Address, error) {
	var address common.Address
	err := repo.statementBuilder.
		Select("hash").
		From(tableNameAddresses).
		Where("id = ?", addressId).
		Scan(&address)
	if err != nil {
		return nil, err
	}
	return &address, nil
}
func (repo *PostgresRepository) GetAddressIdByHash(addressHash common.Address) (storage.AddressId, error) {
	var addressId int64
	err := repo.statementBuilder.
		Select("id").
		From(tableNameAddresses).
		Where(`"hash" = ?`, addressHash).
		Scan(&addressId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return postgresSerialId(addressId), nil
}
func (repo *PostgresRepository) GetReceiptByTransactionId(transactionId storage.TransactionId) (*storage.Receipt, error) {
	var receipt storage.Receipt
	var gasUsed, cumulativeGasUsed, effectiveGasPrice []byte
	var contractAddressId *postgresSerialId
	err := repo.statementBuilder.
		Select(tableColumnsReceipts[1:]...).
		From(tableNameReceipts).
		Where("transaction_id = ?", transactionId).
		Scan(&cumulativeGasUsed,
			&gasUsed,
			&contractAddressId,
			&receipt.PostState,
			&receipt.Status,
			&effectiveGasPrice)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	a := storage.AddressId(contractAddressId)
	receipt.TransactionId = transactionId
	receipt.ContractAddressId = &a
	receipt.GasUsed = bytesToUint64(gasUsed)
	receipt.CumulativeGasUsed = bytesToUint64(cumulativeGasUsed)
	receipt.EffectiveGasPrice = new(big.Int).SetBytes(effectiveGasPrice)
	return &receipt, nil
}
func (repo *PostgresRepository) GetByteCode(bytecodeId storage.BytecodeId) (bytecode *storage.Bytecode, err error) {
	err = repo.statementBuilder.Select("bytecode").From(`"Bytecodes"`).Where("id = ?", bytecodeId).Scan(&bytecode)
	return
}
func (repo *PostgresRepository) GetLogById(logId storage.LogId) (*storage.Log, error) {
	var log storage.Log
	var index []byte
	err := repo.statementBuilder.
		Select(tableColumnsLogs...).
		From(tableNameLogs).
		Where("transaction_id = ?", logId.TransactionId).
		Where(`"index" = ?`, uint64ToBytes(logId.LogIndex)).
		Scan(&log.TransactionId,
			&index,
			&log.AddressId,
			&log.Topic0Id,
			&log.Topic1,
			&log.Topic2,
			&log.Topic3,
			&log.Data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	log.LogIndex = bytesToUint64(index)
	return &log, nil
}
func (repo *PostgresRepository) GetErc20TokenByAddressId(addressId storage.AddressId) (*storage.Erc20Token, error) {
	erc20Token := storage.Erc20Token{
		AddressId: addressId,
	}
	var totalSupply []byte
	err := repo.statementBuilder.
		Select(tableColumnsErc20Tokens[1:]...).
		From(tableNameErc20Tokens).
		Where("address_id = ?", addressId).
		Scan(&erc20Token.Symbol,
			&erc20Token.Name,
			&erc20Token.Decimals,
			&totalSupply)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	erc20Token.TotalSupply = new(big.Int).SetBytes(totalSupply)
	return &erc20Token, err
}
func (repo *PostgresRepository) GetContractByAddressId(addressId storage.AddressId) (*storage.Contract, error) {
	var transactionId, bytecodeId postgresSerialId
	err := repo.statementBuilder.
		Select(tableColumnsContracts[1:]...).
		From(tableNameContracts).
		Where("address_id = ?", addressId).
		Scan(&transactionId, &bytecodeId)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &storage.Contract{
		AddressId:     addressId,
		TransactionId: transactionId,
		BytecodeId:    bytecodeId,
	}, nil
}
func (repo *PostgresRepository) GetLastStoredEtherBalance(addressId storage.AddressId) (*storage.EtherBalance, error) {
	var etherBalance storage.EtherBalance
	var balance []byte
	err := repo.statementBuilder.
		Select("block_id", "balance").
		From(tableNameEtherBalances).
		Where("address_id = ?", addressId).
		OrderBy("block_id DESC").
		Limit(1).Scan(&etherBalance.BlockNumber, &balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	etherBalance.AddressId = addressId
	etherBalance.Balance = new(big.Int).SetBytes(balance)
	return &etherBalance, err
}
func (repo *PostgresRepository) GetLastStoredErc20TokenBalance(addressId, tokenAddressId storage.AddressId) (*storage.Erc20TokenBalance, error) {
	var erc20TokenBalance storage.Erc20TokenBalance
	var balance []byte
	err := repo.statementBuilder.
		Select("block_id", "balance").
		From(tableNameErc20TokenBalances).
		Where("address_id = ?", addressId).
		Where("token_address_id = ?", tokenAddressId).
		OrderBy("block_id DESC").
		Limit(1).
		Scan(&erc20TokenBalance.BlockNumber, &balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	erc20TokenBalance.AddressId = addressId
	erc20TokenBalance.TokenAddressId = tokenAddressId
	erc20TokenBalance.Balance = new(big.Int).SetBytes(balance)
	return &erc20TokenBalance, err
}
func (repo *PostgresRepository) GetLatestBlockNumber() (*storage.BlockNumber, error) {
	var blockNumber *storage.BlockNumber
	err := repo.statementBuilder.Select("MAX(number)").From(tableNameBlocks).Scan(&blockNumber)
	if err != nil {
		return nil, err
	}
	return blockNumber, err
}
func (repo *PostgresRepository) GetWeiBalanceAtBlock(addressId storage.AddressId, blockNumber storage.BlockNumber) (*big.Int, error) {
	var balance []byte
	err := repo.statementBuilder.
		Select("balance").
		From(tableNameEtherBalances).
		Where("address_id = ?", addressId).
		Where("block_id <= ?", blockNumber).
		OrderBy("block_id DESC").
		Limit(1).
		Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return new(big.Int), nil
		}
		return nil, err
	}
	return new(big.Int).SetBytes(balance), nil
}
func (repo *PostgresRepository) GetFirstTxSent(addressId storage.AddressId) (*common.Hash, error) {
	var oldestTxId *uint64
	err := repo.statementBuilder.
		Select("MIN(id)").
		From(tableNameTransactions).
		Where("from_address_id = ?", addressId).Scan(&oldestTxId)
	if err != nil || oldestTxId == nil {
		return nil, err
	}
	firstTxSent, err := repo.GetTransactionById(postgresSerialId(*oldestTxId))
	if err != nil {
		return nil, err
	}
	return &firstTxSent.Hash, nil
}
func (repo *PostgresRepository) GetLastTxSent(addressId storage.AddressId) (*common.Hash, error) {
	var latestTxId *uint64
	err := repo.statementBuilder.
		Select("MAX(id)").
		From(tableNameTransactions).
		Where("from_address_id = ?", addressId).
		Scan(&latestTxId)
	if err != nil || latestTxId == nil {
		return nil, err
	}
	lastTxSent, err := repo.GetTransactionById(postgresSerialId(*latestTxId))
	if err != nil {
		return nil, err
	}
	return &lastTxSent.Hash, nil
}
func (repo *PostgresRepository) GetErc20TokenHolders(erc20TokenId storage.Erc20TokenId) (holders uint64, err error) {
	err = repo.statementBuilder.
		Select("COUNT(DISTINCT token_address_id) AS holder_count").
		From(tableNameErc20TokenBalances).
		Where(`balance > E'\\x00'`).Scan(&holders)
	return
}
func (repo *PostgresRepository) GetUncleByUncleHash(uncleHash *common.Hash) (*storage.Uncle, error) {
	var uncle storage.Uncle
	var minerAddressId postgresSerialId
	var difficulty, reward []byte
	err := repo.statementBuilder.
		Select(tableColumnsUncles...).
		From(tableNameUncles).
		Where("hash = ?", uncleHash).
		Scan(
			&uncle.Hash,
			&uncle.Position,
			&uncle.UncleHeight,
			&uncle.BlockHeight,
			&uncle.ParentHash,
			&minerAddressId,
			&difficulty,
			&uncle.GasLimit,
			&uncle.GasUsed,
			&reward,
			&uncle.Timestamp)
	if err != nil {
		return nil, err
	}
	uncle.MinerAddressId = minerAddressId
	uncle.Difficulty = new(big.Int).SetBytes(difficulty)
	uncle.Reward = new(big.Int).SetBytes(reward)
	return &uncle, nil
}
