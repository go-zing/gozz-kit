package zsql

import (
	"database/sql"
	"testing"
)

func TestName(t *testing.T) {
	s := new(sql.DB)
	t.Log(sessionKey{s} == sessionKey{s})
}
