package postgres

import (
	"database/sql"
	"fmt"
	"math/big"

	sq "github.com/Masterminds/squirrel"
	"github.com/ethereum/go-ethereum/common"
	storage "github.com/librescan-org/backend-db"
)

func (repo *PostgresRepository) ListBlocks(pagination storage.OffsetPagination) (blocks []*storage.Block, totalRecordsFound uint64, err error) {
	if pagination.Limit != 0 {
		var rows *sql.Rows
		rows, err = repo.statementBuilder.
			Select(tableColumnsBlocks...).
			From(tableNameBlocks).
			OrderBy("number DESC").
			Limit(uint64(pagination.Limit)).
			Offset(pagination.Offset).Query()
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			var block *storage.Block
			block, err = scanBlock(rows)
			if err != nil {
				return
			}
			blocks = append(blocks, block)
		}
	}
	err = repo.statementBuilder.Select("COUNT(*)").From(tableNameBlocks).Scan(&totalRecordsFound)
	return
}
func (repo *PostgresRepository) ListUnclesByBlockNumber(blockNumber storage.BlockNumber) (uncles []*storage.Uncle, err error) {
	rows, err := repo.statementBuilder.
		Select(tableColumnsUncles...).
		From(tableNameUncles).
		Where("block_height = ?", blockNumber).
		OrderBy("position").Query()
	if err != nil {
		return
	}
	defer rows.Close()
	var minerAddressId postgresSerialId
	var difficulty, reward []byte
	for rows.Next() {
		var uncle storage.Uncle
		err = rows.Scan(
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
			&uncle.Timestamp,
		)
		if err != nil {
			return
		}
		uncle.MinerAddressId = minerAddressId
		uncle.Difficulty = new(big.Int).SetBytes(difficulty)
		uncle.Reward = new(big.Int).SetBytes(reward)
		uncles = append(uncles, &uncle)
	}
	return
}
func (repo *PostgresRepository) ListTracesByTransactionHash(transactionHash *common.Hash) (traceActions []*storage.TraceAction, blockNumber, timestamp uint64, err error) {
	_, transactionId, err := repo.GetTransactionByHash(transactionHash)
	if err != nil {
		return nil, 0, 0, err
	}
	traces, blockNumbers, timestamps, _, err := repo.listTraces([]any{`transaction_id = ?`, transactionId}, nil)
	if err != nil {
		return nil, 0, 0, err
	}
	if len(traces) != 0 {
		blockNumber = blockNumbers[0]
		timestamp = timestamps[0]
	}
	return traces, blockNumber, timestamp, nil
}
func (repo *PostgresRepository) ListTransactionsByBlockNumber(blockNumber storage.BlockNumber, pagination *storage.OffsetPagination) (transactions []*storage.Transaction, transactionIds []storage.TransactionId, totalRecordsFound uint64, err error) {
	if pagination == nil || pagination.Limit != 0 {
		selectBuilder := repo.statementBuilder.Select(tableColumnsTransactions...).
			From(tableNameTransactions).
			Where("block_id = ?", blockNumber).
			OrderBy(`"index" DESC`)
		if pagination != nil {
			selectBuilder = selectBuilder.
				Limit(uint64(pagination.Limit)).
				Offset(pagination.Offset)
		}
		rows, err := selectBuilder.Query()
		if err != nil {
			return nil, nil, 0, err
		}
		defer rows.Close()
		for rows.Next() {
			tx, txId, err := scanTransaction(rows)
			if err != nil {
				return nil, nil, 0, err
			}
			transactions = append(transactions, tx)
			transactionIds = append(transactionIds, txId)
		}
	}
	err = repo.statementBuilder.
		Select("COUNT(*)").
		From(tableNameTransactions).
		Where("block_id = ?", blockNumber).
		Scan(&totalRecordsFound)
	if err != nil {
		return nil, nil, 0, err
	}
	return
}
func (repo *PostgresRepository) ListTransactionsByAddress(address common.Address, pagination storage.OffsetPagination) (transactions []*storage.Transaction, transactionIds []storage.TransactionId, totalRecordsFound uint64, err error) {
	addressId, err := repo.GetAddressIdByHash(address)
	if err != nil {
		return
	}
	rows, err := repo.statementBuilder.
		Select(tableColumnsTransactions...).
		From(tableNameTransactions).
		Where("? IN (from_address_id, to_address_id)", addressId).
		OrderBy("id DESC").
		Limit(uint64(pagination.Limit)).
		Offset(pagination.Offset).
		Query()
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		tx, txId, err := scanTransaction(rows)
		if err != nil {
			return nil, nil, 0, err
		}
		transactions = append(transactions, tx)
		transactionIds = append(transactionIds, txId)
	}
	err = repo.statementBuilder.
		Select("COUNT(*)").
		From(tableNameTransactions).
		Where("? IN (from_address_id, to_address_id)", addressId).
		Scan(&totalRecordsFound)
	if err != nil {
		return nil, nil, 0, err
	}
	return
}
func (repo *PostgresRepository) ListTransactions(pagination storage.OffsetPagination) (transactions []*storage.Transaction, transactionIds []storage.TransactionId, totalRecordsFound uint64, err error) {
	rows, err := repo.statementBuilder.
		Select(tableColumnsTransactions...).
		From(tableNameTransactions).
		OrderBy("id DESC").
		Limit(uint64(pagination.Limit)).
		Offset(pagination.Offset).
		Query()
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		tx, txId, err := scanTransaction(rows)
		if err != nil {
			return nil, nil, 0, err
		}
		transactions = append(transactions, tx)
		transactionIds = append(transactionIds, txId)
	}
	err = repo.statementBuilder.
		Select("COUNT(*)").
		From(tableNameTransactions).
		Scan(&totalRecordsFound)
	return
}
func (repo *PostgresRepository) ListStorageKeysByTransactionId(transactionId storage.TransactionId) (storageKeys []*storage.StorageKey, err error) {
	rows, err := repo.statementBuilder.
		Select(tableColumnsStorageKeys[1:]...).
		From(tableNameStorageKeys).
		Where("transaction_id = ?", transactionId).
		Query()
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var addressId postgresSerialId
		var storageKeyAsBytes []byte
		if err = rows.Scan(&addressId, &storageKeyAsBytes); err != nil {
			return
		}
		storageKeys = append(storageKeys, &storage.StorageKey{
			TransactionId: transactionId,
			AddressId:     addressId,
			StorageKey:    new(big.Int).SetBytes(storageKeyAsBytes)},
		)
	}
	return
}
func (repo *PostgresRepository) ListLogsByTransactionId(transactionId storage.TransactionId) ([]*storage.Log, error) {
	rows, err := repo.statementBuilder.
		Select(tableColumnsLogs...).
		From(tableNameLogs).
		Where("transaction_id = ?", transactionId).
		OrderBy(`"index" DESC`).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLogs(rows)
}
func (repo *PostgresRepository) ListErc20TokenTransfers(token, fromOrToFilter *common.Address, pagination *storage.OffsetPagination) ([]*storage.Erc20TokenTransfer, uint64, error) {
	filterQuery := func(selectBuilder sq.SelectBuilder) (*sq.SelectBuilder, error) {
		if token != nil {
			addressId, err := repo.GetAddressIdByHash(*token)
			if err != nil {
				return nil, err
			}
			selectBuilder = selectBuilder.Where("token_address_id = ?", addressId)
		}
		if fromOrToFilter != nil {
			addressId, err := repo.GetAddressIdByHash(*fromOrToFilter)
			if err != nil {
				return nil, err
			}
			selectBuilder = selectBuilder.Where("? IN (from_address_id, to_address_id)", addressId)
		}
		return &selectBuilder, nil
	}
	var totalTransferCount uint64
	selectBuilder, err := filterQuery(repo.statementBuilder.
		Select("COUNT(*)").
		From(tableNameErc20TokenTransfers))
	if err != nil {
		return nil, 0, err
	}
	if err = selectBuilder.Scan(&totalTransferCount); err != nil || pagination == nil {
		return nil, totalTransferCount, err
	}
	statement, err := filterQuery(repo.statementBuilder.
		Select(tableColumnsErc20TokenTransfers...).
		From(tableNameErc20TokenTransfers).
		OrderBy("transaction_id DESC", "log_index DESC").
		Limit(uint64(pagination.Limit)).
		Offset(pagination.Offset))
	if err != nil {
		return nil, 0, err
	}
	rows, err := statement.Query()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var transactionId int64
	var logIndex, value []byte
	var tokenAddressId, fromAddressId, toAddressId postgresSerialId
	var transfers []*storage.Erc20TokenTransfer
	for rows.Next() {
		var transfer storage.Erc20TokenTransfer
		err = rows.Scan(
			&transactionId,
			&logIndex,
			&tokenAddressId,
			&fromAddressId,
			&toAddressId,
			&value,
		)
		if err != nil {
			return nil, 0, err
		}
		transfer.LogId = storage.LogId{
			TransactionId: postgresSerialId(transactionId),
			LogIndex:      bytesToUint64(logIndex),
		}
		transfer.TokenAddressId = tokenAddressId
		transfer.FromAddressId = fromAddressId
		transfer.ToAddressId = toAddressId
		transfer.Value = new(big.Int).SetBytes(value)
		transfers = append(transfers, &transfer)
	}
	return transfers, totalTransferCount, err
}
func (repo *PostgresRepository) ListTracesByBlockNumber(blockNumber storage.BlockNumber, pagination *storage.OffsetPagination) ([]*storage.TraceAction, uint64, uint64, error) {
	traces, _, timestamps, totalRecordsFound, err := repo.listTraces([]any{`block_id = ?`, blockNumber}, pagination)
	if err != nil {
		return nil, 0, 0, err
	}
	var timestamp uint64
	if len(timestamps) != 0 {
		timestamp = timestamps[0]
	}
	return traces, timestamp, totalRecordsFound, nil
}
func (repo *PostgresRepository) ListTraces(pagination *storage.OffsetPagination) ([]*storage.TraceAction, []uint64, []uint64, uint64, error) {
	return repo.listTraces(nil, pagination)
}
func (repo *PostgresRepository) listTraces(whereClause []any, pagination *storage.OffsetPagination) ([]*storage.TraceAction, []uint64, []uint64, uint64, error) {
	addFilterLogic := func(selectBuilder sq.SelectBuilder) sq.SelectBuilder {
		mainSelectBuilder := selectBuilder.
			From(tableNameTraces).
			Join(fmt.Sprintf(`%s AS t ON t.id = "transaction_id"`, tableNameTransactions)).
			Join(fmt.Sprintf(`%s AS b ON b.number = t.block_id`, tableNameBlocks))
		if whereClause != nil {
			mainSelectBuilder = mainSelectBuilder.Where(whereClause[0], whereClause[1:]...)
		}
		return mainSelectBuilder
	}
	var totalRecordsFound uint64
	err := addFilterLogic(repo.statementBuilder.Select("COUNT(*)")).Scan(&totalRecordsFound)
	if err != nil {
		return nil, nil, nil, 0, err
	}
	selectBuilder := addFilterLogic(
		repo.statementBuilder.Select(
			"b.number",
			"b.timestamp",
			"transaction_id",
			tableNameTraces+".index",
			"type",
			`"Traces".input`,
			tableNameTraces+".from_address_id",
			tableNameTraces+".to_address_id",
			tableNameTraces+".value",
			tableNameTraces+".gas",
			tableNameTraces+".error")).
		OrderBy("transaction_id DESC", tableNameTraces+".index DESC")
	if pagination != nil {
		selectBuilder = selectBuilder.
			Limit(uint64(pagination.Limit)).
			Offset(pagination.Offset)
	}
	rows, err := selectBuilder.Query()
	if err != nil {
		defer rows.Close()
	}
	var traces []*storage.TraceAction
	var blockNumbers, timestamps []uint64
	for rows.Next() {
		var trace storage.TraceAction
		var transactionId int64
		var fromAddressId, toAddressId int64
		var value []byte
		var blockNumber, timestamp uint64
		err = rows.Scan(&blockNumber, &timestamp, &transactionId, &trace.Index, &trace.Type, &trace.Input, &fromAddressId, &toAddressId, &value, &trace.Gas, &trace.Error)
		if err != nil {
			return nil, nil, nil, 0, err
		}
		trace.TransactionId = postgresSerialId(transactionId)
		trace.From = postgresSerialId(fromAddressId)
		trace.To = postgresSerialId(toAddressId)
		trace.Value = new(big.Int).SetBytes(value)
		traces = append(traces, &trace)
		blockNumbers = append(blockNumbers, blockNumber)
		timestamps = append(timestamps, timestamp)
	}
	return traces, blockNumbers, timestamps, totalRecordsFound, nil
}
func (repo *PostgresRepository) ListErc20TokenBalancesAtBlock(addressId storage.AddressId, blockNumber storage.BlockNumber) (erc20TokenBalances []*storage.Erc20TokenBalance, err error) {
	rows, err := repo.statementBuilder.
		Select("DISTINCT ON (token_address_id) block_id", "token_address_id", "balance").
		From(tableNameErc20TokenBalances).
		Where("address_id = ?", addressId).
		Where("block_id <= ?", blockNumber).
		OrderBy("token_address_id", "block_id DESC").
		Query()
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var erc20TokenBalance storage.Erc20TokenBalance
		var balance []byte
		var tokenAddressId postgresSerialId
		err = rows.Scan(&erc20TokenBalance.BlockNumber, &tokenAddressId, &balance)
		if err != nil {
			return
		}
		erc20TokenBalance.AddressId = addressId
		erc20TokenBalance.TokenAddressId = tokenAddressId
		erc20TokenBalance.Balance = new(big.Int).SetBytes(balance)
		erc20TokenBalances = append(erc20TokenBalances, &erc20TokenBalance)
	}
	return
}
func (repo *PostgresRepository) ListStateChangesByTransactionHash(transactionHash *common.Hash) ([]*storage.StateChange, error) {
	transaction, transactionId, err := repo.GetTransactionByHash(transactionHash)
	if err != nil {
		return nil, err
	}
	if transaction == nil {
		return nil, nil
	}
	stateChangesRows, err := repo.statementBuilder.
		Select(tableColumnsStateChanges[1:]...).
		From(tableNameStateChanges).
		Where("transaction_id = ?", transactionId).
		Query()
	if err != nil {
		return nil, err
	}
	var stateChanges []*storage.StateChange
	for stateChangesRows.Next() {
		var addressId postgresSerialId
		var balanceBefore, balanceAfter []byte
		var nullableNonceBefore, nullableNonceAfter *[]byte
		if err = stateChangesRows.Scan(&addressId, &balanceBefore, &balanceAfter, &nullableNonceBefore, &nullableNonceAfter); err != nil {
			stateChangesRows.Close()
			return nil, err
		}
		var nonceBefore *uint64
		if nullableNonceBefore != nil {
			n := bytesToUint64(*nullableNonceBefore)
			nonceBefore = &n
		}
		var nonceAfter *uint64
		if nullableNonceBefore != nil {
			n := bytesToUint64(*nullableNonceAfter)
			nonceAfter = &n
		}
		stateChanges = append(stateChanges, &storage.StateChange{
			TransactionId: transactionId,
			AddressId:     addressId,
			BalanceBefore: new(big.Int).SetBytes(balanceBefore),
			BalanceAfter:  new(big.Int).SetBytes(balanceAfter),
			NonceBefore:   nonceBefore,
			NonceAfter:    nonceAfter,
		})
	}
	if err = stateChangesRows.Close(); err != nil {
		return nil, err
	}
	for _, stateChange := range stateChanges {
		storageChangesRows, err := repo.statementBuilder.
			Select(tableColumnsStorageChanges[1:]...).
			From(tableNammeStorageChanges).
			Where("transaction_id = ?", stateChange.TransactionId).
			Where("address_id = ?", stateChange.AddressId).
			Query()
		if err != nil {
			return nil, err
		}
		var storageChanges []*storage.StorageChange
		for storageChangesRows.Next() {
			var addressId postgresSerialId
			var storageAddress, valueBefore, valueAfter []byte
			if err = storageChangesRows.Scan(&addressId, &storageAddress, &valueBefore, &valueAfter); err != nil {
				storageChangesRows.Close()
				return nil, err
			}
			storageChanges = append(storageChanges, &storage.StorageChange{
				TransactionId:  transactionId,
				AddressId:      addressId,
				StorageAddress: new(big.Int).SetBytes(storageAddress),
				ValueBefore:    new(big.Int).SetBytes(valueBefore),
				ValueAfter:     new(big.Int).SetBytes(valueAfter),
			})
		}
		storageChangesRows.Close()
		stateChange.StorageChanges = storageChanges
	}

	return stateChanges, nil
}
