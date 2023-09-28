package postgres

import (
	"crypto/sha256"
	"database/sql"
	"math/big"

	sq "github.com/Masterminds/squirrel"
	"github.com/ethereum/go-ethereum/common"
)

const on_conflict_do_nothing = "ON CONFLICT DO NOTHING"

func executeBulkInsertIgnore[Record any, Records []Record](tx *sql.Tx, records Records, insertQueryBuilder func(Record) sq.InsertBuilder) (err error) {
	for _, record := range records {
		if _, err = insertQueryBuilder(record).RunWith(tx).Suffix(on_conflict_do_nothing).Exec(); err != nil {
			break
		}
	}
	return
}
func bigIntMustNotBeNil(number *big.Int) *big.Int {
	if number == nil {
		return new(big.Int)
	}
	return number
}
func (repo *PostgresRepository) StoreAddress(addresses ...common.Address) (addressIds []storage.AddressId, err error) {
	for _, address := range addresses {
		_, err = repo.statementBuilder.
			Insert(tableNameAddresses).
			Columns("hash").
			Values(address).
			Suffix(on_conflict_do_nothing).
			Exec()
		if err != nil {
			return nil, err
		}
	}
	for _, address := range addresses {
		var id postgresSerialId
		err = repo.statementBuilder.
			Select("id").
			From(tableNameAddresses).
			Where("hash = ?", address).
			Scan(&id)
		if err != nil {
			return nil, err
		}
		addressIds = append(addressIds, id)
	}
	return addressIds, nil
}

func (repo *PostgresRepository) StoreBlock(blocks ...*storage.Block) error {
	return executeBulkInsertIgnore(repo.tx, blocks,
		func(block *storage.Block) sq.InsertBuilder {
			block.Difficulty = bigIntMustNotBeNil(block.Difficulty)
			block.TotalDifficulty = bigIntMustNotBeNil(block.TotalDifficulty)
			block.BaseFeePerGas = bigIntMustNotBeNil(block.BaseFeePerGas)
			block.StaticReward = bigIntMustNotBeNil(block.StaticReward)
			return repo.statementBuilder.
				Insert(tableNameBlocks).
				Columns(tableColumnsBlocks...).
				Values(block.Hash,
					block.Number,
					uint64ToBytes(block.Nonce),
					block.Sha3Uncles,
					block.LogsBloom.Bytes(),
					block.StateRoot,
					block.MinerAddressId,
					block.Difficulty.Bytes(),
					block.TotalDifficulty.Bytes(),
					uint64ToBytes(block.Size),
					block.ExtraData,
					uint64ToBytes(block.GasLimit),
					uint64ToBytes(block.GasUsed),
					block.BaseFeePerGas.Bytes(),
					block.MixHash,
					block.StaticReward.Bytes(),
					block.Timestamp)
		})
}
func (repo *PostgresRepository) StoreUncle(uncles ...*storage.Uncle) error {
	return executeBulkInsertIgnore(repo.tx, uncles,
		func(uncle *storage.Uncle) sq.InsertBuilder {
			uncle.Difficulty = bigIntMustNotBeNil(uncle.Difficulty)
			uncle.Reward = bigIntMustNotBeNil(uncle.Reward)
			return repo.statementBuilder.
				Insert(tableNameUncles).
				Columns(tableColumnsUncles...).
				Values(
					uncle.Hash,
					uncle.Position,
					uncle.UncleHeight,
					uncle.BlockHeight,
					uncle.ParentHash.Bytes(),
					uncle.MinerAddressId,
					uncle.Difficulty.Bytes(),
					uncle.GasLimit,
					uncle.GasUsed,
					uncle.Reward.Bytes(),
					uncle.Timestamp)
		})
}
func (repo *PostgresRepository) StoreTransaction(transactions ...*storage.Transaction) ([]storage.TransactionId, error) {
	err := executeBulkInsertIgnore(repo.tx, transactions, func(transaction *storage.Transaction) sq.InsertBuilder {
		var nullableGasTipCap, nullableGasFeeCap any
		if transaction.GasTipCap == nil {
			nullableGasTipCap = "NULL"
		} else {
			nullableGasTipCap = transaction.GasTipCap.Bytes()
		}
		if transaction.GasFeeCap == nil {
			nullableGasFeeCap = "NULL"
		} else {
			nullableGasFeeCap = transaction.GasFeeCap.Bytes()
		}
		return repo.statementBuilder.
			Insert(tableNameTransactions).
			Columns(tableColumnsTransactions[1:]...).
			Values(transaction.BlockNumber,
				transaction.Hash,
				uint64ToBytes(transaction.Nonce),
				uint64ToBytes(transaction.Index),
				transaction.FromAddressId,
				transaction.ToAddressId,
				transaction.Value.Bytes(),
				uint64ToBytes(transaction.Gas),
				transaction.GasPrice.Bytes(),
				nullableGasTipCap,
				nullableGasFeeCap,
				transaction.Input,
				transaction.Type)
	})
	if err != nil {
		return nil, err
	}
	var transactionIds []storage.TransactionId
	for _, transaction := range transactions {
		var id postgresSerialId
		err = repo.statementBuilder.
			Select("id").
			From(tableNameTransactions).
			Where("hash = ?", transaction.Hash).
			Scan(&id)
		if err != nil {
			return nil, err
		}
		transactionIds = append(transactionIds, id)
	}
	return transactionIds, nil
}
func (repo *PostgresRepository) StoreStorageKey(storageKeys ...*storage.StorageKey) error {
	return executeBulkInsertIgnore(repo.tx, storageKeys, func(storageKey *storage.StorageKey) sq.InsertBuilder {
		return repo.statementBuilder.
			Insert(tableNameStorageKeys).
			Columns(tableColumnsStorageKeys...).
			Values(storageKey.TransactionId,
				storageKey.AddressId,
				storageKey.StorageKey.Bytes())
	})
}
func (repo *PostgresRepository) StoreReceipt(receipts ...*storage.Receipt) error {
	return executeBulkInsertIgnore(repo.tx, receipts,
		func(receipt *storage.Receipt) sq.InsertBuilder {
			return repo.statementBuilder.
				Insert(tableNameReceipts).
				Columns(tableColumnsReceipts...).
				Values(receipt.TransactionId,
					uint64ToBytes(receipt.CumulativeGasUsed),
					uint64ToBytes(receipt.GasUsed),
					receipt.ContractAddressId,
					receipt.PostState.Bytes(),
					receipt.Status,
					receipt.EffectiveGasPrice.Bytes())
		})
}
func (repo *PostgresRepository) StoreLog(logs ...*storage.Log) error {
	return executeBulkInsertIgnore(repo.tx, logs, func(log *storage.Log) sq.InsertBuilder {
		var topic1, topic2, topic3 []byte
		if log.Topic1 != nil {
			topic1 = log.Topic1[:]
		}
		if log.Topic2 != nil {
			topic2 = log.Topic2[:]
		}
		if log.Topic3 != nil {
			topic3 = log.Topic3[:]
		}
		return repo.statementBuilder.
			Insert(tableNameLogs).
			Columns(tableColumnsLogs...).
			Values(log.TransactionId,
				uint64ToBytes(log.LogIndex),
				log.AddressId,
				log.Topic0Id,
				topic1,
				topic2,
				topic3,
				log.Data)
	})
}
func (repo *PostgresRepository) StoreTopic0(eventTypes ...*storage.EventType) ([]storage.Topic0Id, error) {
	err := executeBulkInsertIgnore(repo.tx, eventTypes, func(eventType *storage.EventType) sq.InsertBuilder {
		return repo.statementBuilder.
			Insert(tableNameEventTypes).
			Columns(tableColumnsEventTypes[1:]...).
			Values(eventType.Hash, eventType.Signature)
	})
	if err != nil {
		return nil, err
	}
	var eventTypeIds []storage.Topic0Id
	for _, eventType := range eventTypes {
		var id postgresSerialId
		err = repo.statementBuilder.Select("id").From(tableNameEventTypes).Where("hash = ?", eventType.Hash).Scan(&id)
		if err != nil {
			return nil, err
		}
		eventTypeIds = append(eventTypeIds, postgresSerialId(id))
	}
	return eventTypeIds, nil
}
func (repo *PostgresRepository) StoreErc20TokenTransfer(erc20TokenTransfers ...*storage.Erc20TokenTransfer) error {
	return executeBulkInsertIgnore(repo.tx, erc20TokenTransfers,
		func(erc20TokenTransfer *storage.Erc20TokenTransfer) sq.InsertBuilder {

			return repo.statementBuilder.
				Insert(tableNameErc20TokenTransfers).
				Columns(tableColumnsErc20TokenTransfers...).
				Values(
					erc20TokenTransfer.LogId.TransactionId,
					uint64ToBytes(erc20TokenTransfer.LogId.LogIndex),
					erc20TokenTransfer.TokenAddressId,
					erc20TokenTransfer.FromAddressId,
					erc20TokenTransfer.ToAddressId,
					erc20TokenTransfer.Value.Bytes())
		})
}
func (repo *PostgresRepository) StoreErc20Token(erc20Tokens ...*storage.Erc20Token) error {
	return executeBulkInsertIgnore(repo.tx, erc20Tokens,
		func(erc20Token *storage.Erc20Token) sq.InsertBuilder {
			return repo.statementBuilder.
				Insert(tableNameErc20Tokens).
				Columns(tableColumnsErc20Tokens...).
				Values(erc20Token.AddressId,
					erc20Token.Symbol,
					erc20Token.Name,
					erc20Token.Decimals,
					erc20Token.TotalSupply.Bytes())
		})
}
func (repo *PostgresRepository) StoreContract(contracts ...*storage.Contract) error {
	return executeBulkInsertIgnore(repo.tx, contracts,
		func(contract *storage.Contract) sq.InsertBuilder {
			return repo.statementBuilder.
				Insert(tableNameContracts).
				Columns(tableColumnsContracts...).
				Values(contract.AddressId, contract.TransactionId, contract.BytecodeId)
		})
}
func (repo *PostgresRepository) StoreBytecode(bytecodes ...storage.Bytecode) ([]storage.BytecodeId, error) {
	if len(bytecodes) != 1 {
		panic("bulk insert not implemented")
	}
	bytecode := bytecodes[0]
	sha256Digest := sha256.Sum256(bytecode)
	hash := sha256Digest[:]
	_, err := repo.statementBuilder.
		Insert(tableNameBytecodes).
		Columns(tableColumnsBytecodes...).
		Values(bytecode, hash).
		Suffix(on_conflict_do_nothing).
		Exec()
	if err != nil {
		return nil, err
	}
	var id postgresSerialId
	err = repo.statementBuilder.
		Select("id").
		From(tableNameBytecodes).
		Where("sha256 = ?", hash).
		Scan(&id)
	if err != nil {
		return nil, err
	}
	return []storage.BytecodeId{id}, nil
}
func (repo *PostgresRepository) StoreTrace(traceActions ...*storage.TraceAction) error {
	return executeBulkInsertIgnore(repo.tx, traceActions,
		func(traceAction *storage.TraceAction) sq.InsertBuilder {
			return repo.statementBuilder.
				Insert(tableNameTraces).
				Columns(tableColumnsTraces...).
				Values(traceAction.TransactionId,
					traceAction.Index,
					traceAction.Type,
					traceAction.Input,
					traceAction.From,
					traceAction.To,
					traceAction.Value.Bytes(),
					traceAction.Gas,
					traceAction.Error)
		})
}
func (repo *PostgresRepository) StoreEtherBalance(etherBalances ...*storage.EtherBalance) error {
	return executeBulkInsertIgnore(repo.tx, etherBalances,
		func(etherBalance *storage.EtherBalance) sq.InsertBuilder {
			return repo.statementBuilder.
				Insert(tableNameEtherBalances).
				Columns(tableColumnsEtherBalances...).
				Values(etherBalance.AddressId, etherBalance.BlockNumber, etherBalance.Balance.Bytes())
		})
}
func (repo *PostgresRepository) StoreErc20TokenBalance(tokenBalances ...*storage.Erc20TokenBalance) error {
	return executeBulkInsertIgnore(repo.tx, tokenBalances,
		func(tokenBalance *storage.Erc20TokenBalance) sq.InsertBuilder {
			return repo.statementBuilder.
				Insert(tableNameErc20TokenBalances).
				Columns(tableColumnsErc20TokenBalances...).
				Values(tokenBalance.AddressId, tokenBalance.BlockNumber, tokenBalance.TokenAddressId, tokenBalance.Balance.Bytes())
		})
}
func (repo *PostgresRepository) StoreStateChange(stateChanges ...*storage.StateChange) error {
	return executeBulkInsertIgnore(repo.tx, stateChanges, func(stateChange *storage.StateChange) sq.InsertBuilder {
		var nullableBalanceBefore any
		if stateChange.BalanceAfter == nil {
			nullableBalanceBefore = "NULL"
		} else {
			nullableBalanceBefore = stateChange.BalanceBefore.Bytes()
		}
		var nullableNonceBefore any
		if stateChange.NonceBefore == nil {
			nullableNonceBefore = "NULL"
		} else {
			nullableNonceBefore = uint64ToBytes(*stateChange.NonceBefore)
		}
		var nullableNonceAfter any
		if stateChange.NonceAfter == nil {
			nullableNonceAfter = "NULL"
		} else {
			nullableNonceAfter = uint64ToBytes(*stateChange.NonceAfter)
		}
		return repo.statementBuilder.
			Insert(tableNameStateChanges).
			Columns(tableColumnsStateChanges...).
			Values(
				stateChange.TransactionId,
				stateChange.AddressId,
				nullableBalanceBefore,
				bigIntMustNotBeNil(stateChange.BalanceAfter).Bytes(),
				nullableNonceBefore,
				nullableNonceAfter)
	})
}
func (repo *PostgresRepository) StoreStorageChange(storageChanges ...*storage.StorageChange) error {
	return executeBulkInsertIgnore(repo.tx, storageChanges, func(storageChange *storage.StorageChange) sq.InsertBuilder {
		return repo.statementBuilder.
			Insert(tableNammeStorageChanges).
			Columns(tableColumnsStorageChanges...).
			Values(
				storageChange.TransactionId,
				storageChange.AddressId,
				storageChange.StorageAddress.Bytes(),
				storageChange.ValueBefore.Bytes(),
				storageChange.ValueAfter.Bytes())
	})
}
