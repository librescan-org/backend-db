package postgres

import (
	"database/sql"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	storage "github.com/librescan-org/backend-db"
)

type ScannerWithErrHandling interface {
	Scan(...any) error
	Err() error
}

func scanBlock(scanner ScannerWithErrHandling) (*storage.Block, error) {
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	var block = storage.Block{}
	var nonce, logsBlooms, difficulty, totalDifficulty, size, gasLimit, gasUsed, staticReward, baseFeePerGas []byte

	var minerAddressId postgresSerialId
	err := scanner.Scan(
		&block.Hash,
		&block.Number,
		&nonce,
		&block.Sha3Uncles,
		&logsBlooms,
		&block.StateRoot,
		&minerAddressId,
		&difficulty,
		&totalDifficulty,
		&size,
		&block.ExtraData,
		&gasLimit,
		&gasUsed,
		&baseFeePerGas,
		&block.MixHash,
		&staticReward,
		&block.Timestamp)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err == nil {
		block.Nonce = bytesToUint64(nonce)
		block.LogsBloom = types.BytesToBloom(logsBlooms)
		block.MinerAddressId = minerAddressId
		block.StaticReward = new(big.Int).SetBytes(staticReward)
		block.Difficulty = new(big.Int).SetBytes(difficulty)
		block.TotalDifficulty = new(big.Int).SetBytes(totalDifficulty)
		block.Size = bytesToUint64(size)
		block.GasLimit = bytesToUint64(gasLimit)
		block.GasUsed = bytesToUint64(gasUsed)
		block.BaseFeePerGas = new(big.Int).SetBytes(baseFeePerGas)
	}
	return &block, err
}
func scanTransaction(scanner ScannerWithErrHandling) (*storage.Transaction, storage.TransactionId, error) {
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	transaction := &storage.Transaction{}
	var nonce, index, value, gas, gasPrice, gasTipCap, gasFeeCap []byte
	var transactionId, fromAddressId postgresSerialId
	var nullableToAddressId sql.NullInt64
	err := scanner.Scan(
		&transactionId,
		&transaction.BlockNumber,
		&transaction.Hash,
		&nonce,
		&index,
		&fromAddressId,
		&nullableToAddressId,
		&value,
		&gas,
		&gasPrice,
		&gasTipCap,
		&gasFeeCap,
		&transaction.Input,
		&transaction.Type,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	transaction.FromAddressId = fromAddressId
	if nullableToAddressId.Valid {
		id := storage.AddressId(nullableToAddressId.Int64)
		transaction.ToAddressId = &id
	}
	transaction.Nonce = bytesToUint64(nonce)
	transaction.Index = bytesToUint64(index)
	transaction.Value = new(big.Int).SetBytes(value)
	transaction.Gas = bytesToUint64(gas)
	transaction.GasPrice = new(big.Int).SetBytes(gasPrice)
	transaction.GasTipCap = new(big.Int).SetBytes(gasTipCap)
	transaction.GasFeeCap = new(big.Int).SetBytes(gasFeeCap)
	return transaction, transactionId, nil
}
func scanLogs(rows *sql.Rows) (logs []*storage.Log, err error) {
	var transactionId postgresSerialId
	var addressId postgresSerialId
	var topic0Id *any
	var index, topic1, topic2, topic3 []byte
	for rows.Next() {
		var log storage.Log
		err = rows.Scan(
			&transactionId,
			&index,
			&addressId,
			&topic0Id,
			&topic1,
			&topic2,
			&topic3,
			&log.Data)
		if err != nil {
			return
		}
		log.TransactionId = transactionId
		log.AddressId = addressId
		log.Topic0Id = topic0Id
		log.LogIndex = bytesToUint64(index)
		if len(topic1) != 0 {
			log.Topic1 = (*[32]byte)(topic1)
		}
		if len(topic2) != 0 {
			log.Topic2 = (*[32]byte)(topic2)
		}
		if len(topic3) != 0 {
			log.Topic3 = (*[32]byte)(topic3)
		}
		logs = append(logs, &log)
	}
	return
}
