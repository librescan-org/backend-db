package storage

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type LogIndex = uint64
type (
	AddressId       = any
	TransactionId   = any
	ReceiptId       = any
	TokenTransferId = any
	Erc20TokenId    = any
	BytecodeId      = any
	Topic0Id        = any
	LogId           struct {
		TransactionId
		LogIndex
	}
)

// OffsetPagination represents an offset based pagination.
//   - Limit tells the number of records to return.
//   - Offset tells the cursor position from where to start listing.
type OffsetPagination struct {
	Limit  uint8
	Offset uint64
}

// NewOffsetPagination is a convenience method for simple pagination
// using page numbers instead of database-like offsetting.
func NewOffsetPagination(limit uint8, page uint64) OffsetPagination {
	return OffsetPagination{
		Limit:  limit,
		Offset: uint64(limit) * page,
	}
}

type Storage interface {
	Loader
	DataStore
}

// DataStore is the abscract repository managing blockchain data.
// All methods inside this interface MUST use a single, common transaction!
// For finalizing the common transaction, see Loader interface's Commit method's documentation.
type DataStore interface {
	Reader
	Inserter
	Deleter
}
type Loader interface {
	Load() error

	// Commit MUST write all Inserter and Deleter methods's changes.
	// This MUST be the only method responsible for executing all previous Inserter and Deleter operations.
	Commit(context.Context) error
}

// Inserter defines all inserter methods.
// Bulk insertion MUST be supported for all methods.
// All methods MUST execute into a single, common transaction for DataStore, and never commit!
// See Loader's Commit method for explanation.
type Inserter interface {
	StoreAddress(...common.Address) ([]AddressId, error)
	StoreBlock(...*Block) error
	StoreUncle(...*Uncle) error
	StoreTransaction(...*Transaction) ([]TransactionId, error)
	StoreStorageKey(...*StorageKey) error
	StoreReceipt(...*Receipt) error
	StoreLog(...*Log) error
	StoreErc20Token(...*Erc20Token) error
	StoreErc20TokenTransfer(...*Erc20TokenTransfer) error
	StoreBytecode(...Bytecode) ([]BytecodeId, error)
	StoreContract(...*Contract) error
	StoreTopic0(...*EventType) ([]Topic0Id, error)
	StoreTrace(...*TraceAction) error
	StoreEtherBalance(...*EtherBalance) error
	StoreErc20TokenBalance(...*Erc20TokenBalance) error
	StoreStateChange(...*StateChange) error
	StoreStorageChange(...*StorageChange) error
}

// All methods MUST execute into a single, common transaction for DataStore, and never commit!
type Deleter interface {
	// DeleteBlockAndAllReferences MUST delete all blocks having the specified block numbers, including all other entity references in those blocks, recursively.
	// In other words, all insert methods must be undone for the specified block numbers.
	DeleteBlockAndAllReferences(...BlockNumber) error
}

// Reader defines all reader methods.
//
// Lister methods:
//   - Names start with the word "List", and MUST return a list of the requested entities.
//   - Returned entities are pointer types in order to improve efficiency, and they CANNOT be nil.
//   - The slice itself can be nil, or empty slice, they are treated the same way.
//
// Getter methods:
//   - Names start with the word "Get", and MUST return a single entity, OR nil when not found.
//   - If the entity is not found, nil MUST be returned without error, as that is a valid query result.
//
// Most listers accept an OffsetPagination or *OffsetPagination:
//   - OffsetPagination type is mandatory to provide. Even if its Limit parameter is 0, non-list return values MUST still be returned.
//   - *OffsetPagination type is NOT mandatory to provide, can be nil, in which case all entities MUST be returned.
type Reader interface {
	ListBlocks(OffsetPagination) (_ []*Block, totalRecordsFound uint64, _ error)
	ListUnclesByBlockNumber(BlockNumber) ([]*Uncle, error)
	ListTransactions(OffsetPagination) (_ []*Transaction, _ []TransactionId, totalRecordsFound uint64, _ error)
	ListTransactionsByBlockNumber(BlockNumber, *OffsetPagination) (_ []*Transaction, _ []TransactionId, totalRecordsFound uint64, _ error)
	ListTransactionsByAddress(common.Address, OffsetPagination) (_ []*Transaction, _ []TransactionId, totalRecordsFound uint64, _ error)
	ListStorageKeysByTransactionId(TransactionId) ([]*StorageKey, error)
	ListLogsByTransactionId(TransactionId) ([]*Log, error)
	ListErc20TokenTransfers(token, fromOrToFilter *common.Address, _ *OffsetPagination) (_ []*Erc20TokenTransfer, totalRecordsFound uint64, _ error)
	ListTraces(*OffsetPagination) (_ []*TraceAction, blockNumbers []uint64, timestamps []uint64, totalRecordsFound uint64, _ error)
	ListTracesByTransactionHash(transactionHash *common.Hash) (_ []*TraceAction, blockNumber uint64, timestamp uint64, _ error)
	ListTracesByBlockNumber(BlockNumber, *OffsetPagination) (_ []*TraceAction, timestamp uint64, totalRecordsFound uint64, _ error)
	ListErc20TokenBalancesAtBlock(holder AddressId, _ BlockNumber) ([]*Erc20TokenBalance, error)
	ListStateChangesByTransactionHash(*common.Hash) ([]*StateChange, error)
	GetAddressById(AddressId) (*common.Address, error)
	GetAddressIdByHash(common.Address) (AddressId, error)
	GetBlockByHash(*common.Hash) (*Block, error)
	GetBlockByNumber(uint64) (*Block, error)
	GetLatestBlockNumber() (*BlockNumber, error)
	GetUncleByUncleHash(*common.Hash) (*Uncle, error)
	GetTransactionById(TransactionId) (*Transaction, error)
	GetTransactionByHash(*common.Hash) (*Transaction, TransactionId, error)
	GetReceiptByTransactionId(TransactionId) (*Receipt, error)
	GetByteCode(BytecodeId) (*Bytecode, error)
	GetLogById(LogId) (*Log, error)
	GetErc20TokenByAddressId(AddressId) (*Erc20Token, error)
	GetContractByAddressId(AddressId) (*Contract, error)
	GetEventTypeById(Topic0Id) (*EventType, error)
	GetWeiBalanceAtBlock(AddressId, BlockNumber) (*big.Int, error)
	GetLastStoredEtherBalance(AddressId) (*EtherBalance, error)
	GetLastStoredErc20TokenBalance(holder, tokenAddressId AddressId) (*Erc20TokenBalance, error)
	GetFirstTxSent(AddressId) (*common.Hash, error)
	GetLastTxSent(AddressId) (*common.Hash, error)
	GetErc20TokenHolders(Erc20TokenId) (uint64, error)
}
type StateChange struct {
	TransactionId
	AddressId
	BalanceBefore  *big.Int
	BalanceAfter   *big.Int
	NonceBefore    *uint64
	NonceAfter     *uint64
	StorageChanges []*StorageChange
}
type StorageChange struct {
	TransactionId
	AddressId
	StorageAddress *big.Int
	ValueBefore    *big.Int
	ValueAfter     *big.Int
}
type Bytecode []byte
type BlockNumber = uint64
type TokenSymbol string
type EtherBalance struct {
	BlockNumber
	AddressId
	Balance *big.Int
}
type Erc20TokenBalance struct {
	BlockNumber
	AddressId      AddressId
	TokenAddressId AddressId
	Balance        *big.Int
}
type Uncle struct {
	Position    uint8
	Hash        common.Hash
	UncleHeight BlockNumber
	BlockHeight BlockNumber
	ParentHash  common.Hash
	// Sha3Uncles     common.Hash
	MinerAddressId AddressId
	Difficulty     *big.Int
	GasLimit       uint64
	GasUsed        uint64
	Timestamp      uint64
	Reward         *big.Int
}
type Block struct {
	Hash            common.Hash
	Number          BlockNumber
	Nonce           uint64
	Sha3Uncles      common.Hash
	LogsBloom       types.Bloom
	StateRoot       common.Hash
	MinerAddressId  AddressId
	Difficulty      *big.Int
	TotalDifficulty *big.Int
	Size            uint64
	ExtraData       []byte
	GasLimit        uint64
	GasUsed         uint64
	Timestamp       uint64
	BaseFeePerGas   *big.Int
	MixHash         common.Hash
	StaticReward    *big.Int
}
type StorageKey struct {
	TransactionId
	AddressId
	StorageKey *big.Int
}
type Transaction struct {
	BlockNumber
	Hash          common.Hash
	Nonce         uint64
	Index         uint64
	FromAddressId AddressId
	ToAddressId   *AddressId
	Value         *big.Int
	Gas           uint64
	GasPrice      *big.Int
	GasTipCap     *big.Int
	GasFeeCap     *big.Int
	Input         []byte
	Type          uint8
}
type ReceiptStatus = bool

const (
	ReceiptStatusFailure ReceiptStatus = false
	ReceiptStatusSuccess ReceiptStatus = true
)

type Receipt struct {
	TransactionId
	CumulativeGasUsed uint64
	GasUsed           uint64
	ContractAddressId *AddressId
	PostState         common.Hash
	Status            ReceiptStatus
	EffectiveGasPrice *big.Int
}

func (receipt *Receipt) TransactionFee() *big.Int {
	return new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), receipt.EffectiveGasPrice)
}

type Log struct {
	LogId
	AddressId
	Topic0Id *Topic0Id
	Topic1   *[32]byte
	Topic2   *[32]byte
	Topic3   *[32]byte
	Data     []byte
}

type Erc20TokenTransfer struct {
	LogId
	TokenAddressId AddressId
	FromAddressId  AddressId
	ToAddressId    AddressId
	Value          *big.Int
}

type Contract struct {
	AddressId
	TransactionId
	BytecodeId
}
type Erc20Token struct {
	AddressId
	Symbol      string
	Name        string
	Decimals    uint8
	TotalSupply *big.Int
}

type EventType struct {
	common.Hash
	Signature *string
}
type TraceAction struct {
	TransactionId
	Index uint16 // position in the returned traces list
	Type  string
	Input []byte
	From  AddressId
	To    AddressId
	Value *big.Int
	Gas   uint64
	Error *string
}
