package upload

import (
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/maedu/mtg-cards/user/db"
)

type ParseRequest struct {
	UserID string
	Source string
}

type handler interface {
	parse(*csv.Reader, ParseRequest) (bool, []*db.UserCard, error)
	name() string
}

type fields []string

func (fields fields) value(field string, record []string) string {
	for p, v := range fields {
		if v == field {
			return strings.TrimSpace(record[p])
		}
	}
	return ""
}

func rowMatchesFields(row []string, fields fields) bool {
	// If one is nil, the other must also be nil.
	if (row == nil) != (fields == nil) {
		return false
	}

	if len(row) != len(fields) {
		return false
	}

	for i := range row {
		if strings.TrimSpace(row[i]) != fields[i] {
			return false
		}
	}
	return true
}

func parseQuantity(amount string) (int64, error) {
	val := strings.ReplaceAll(amount, "'", "")
	val = strings.ReplaceAll(val, ",", "")
	val = strings.TrimSpace(val)
	return strconv.ParseInt(val, 10, 64)
}

func allFieldsEmpty(record []string) bool {
	for _, v := range record {
		if v != "" {
			return false
		}
	}

	return true
}
