package postgres

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"math/big"
	"os"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

const env_POSTGRES_DB = "POSTGRES_DB"

//go:embed postgres_init.sql
var postgres_init_sql string

// bytesToUint64 converts a byte slice to uint64. Panics on data loss.
func bytesToUint64(bytes []byte) uint64 {
	if len(bytes) > 8 {
		panic("programming error")
	}
	return new(big.Int).SetBytes(bytes).Uint64()
}

type PostgresRepository struct {
	conn             *sql.DB
	tx               *sql.Tx
	statementBuilder sq.StatementBuilderType
}
type postgresSerialId int64

func isErrCode(err error, code pq.ErrorCode) bool {
	if errToAnalyze, ok := err.(*pq.Error); ok {
		return errToAnalyze.Code == code
	}
	return false
}

// uint64ToBytes converts a uint64 to a postgres bytea as efficiently as possible,
// meaning that the result can be a zero length slice, or a maximum of 8 bytes.
func uint64ToBytes(value uint64) []byte {
	return new(big.Int).SetUint64(value).Bytes()
}
func (repo *PostgresRepository) openPostgresConnection(useDefaultDb bool) (err error) {
	var dbName string
	if useDefaultDb {
		dbName = "postgres"
	} else {
		dbName = os.Getenv(env_POSTGRES_DB)
	}
	repo.conn, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		dbName))
	return
}
func (repo *PostgresRepository) Load() (err error) {
	const err_invalid_catalog_name = "3D000"
	databaseName := strings.ToLower(os.Getenv(env_POSTGRES_DB))
	if err = repo.openPostgresConnection(false); err != nil {
		return
	}
	_, err = repo.conn.Exec(postgres_init_sql)
	if isErrCode(err, err_invalid_catalog_name) {
		if err = repo.openPostgresConnection(true); err != nil {
			return
		}
		if _, err = repo.conn.Exec("CREATE DATABASE " + databaseName); err != nil {
			return
		}
		if err = repo.conn.Close(); err != nil {
			return
		}
		if err = repo.openPostgresConnection(false); err == nil {
			_, err = repo.conn.Exec(postgres_init_sql)
		}
	}
	if err == nil {
		repo.tx, err = repo.conn.Begin()
		if err != nil {
			return err
		}
		repo.statementBuilder = sq.StatementBuilder.RunWith(repo.tx).PlaceholderFormat(sq.Dollar)
	}
	return
}
func (repo *PostgresRepository) Commit(ctx context.Context) (err error) {
	err = repo.tx.Commit()
	if err != nil {
		return
	}
	repo.tx, err = repo.conn.BeginTx(ctx, nil)
	if err == nil {
		repo.statementBuilder = repo.statementBuilder.RunWith(repo.tx)
	}
	return
}
