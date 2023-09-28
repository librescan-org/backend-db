package postgres

import (
	"fmt"
	"strconv"
	"strings"

	storage "github.com/librescan-org/backend-db"
)

func (repo *PostgresRepository) DeleteBlockAndAllReferences(blockNumbers ...storage.BlockNumber) error {
	blockNumbersStr := make([]string, 0, len(blockNumbers))
	for _, blockNumber := range blockNumbers {
		blockNumbersStr = append(blockNumbersStr, strconv.FormatUint(blockNumber, 10))
	}
	_, err := repo.statementBuilder.
		Delete(tableNameBlocks).
		Where(fmt.Sprintf("number IN (%s)", strings.Join(blockNumbersStr, ","))).
		Exec()
	return err
}
