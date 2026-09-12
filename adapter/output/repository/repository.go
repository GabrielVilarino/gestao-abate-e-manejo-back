package repository

import (
	"database/sql"
	"fmt"
)

func ensureDataFound(result sql.Result, id int) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("registro com id %d não encontrado", id)
	}
	return nil
}
