package excel_adapter

import (
	"strings"

	"github.com/fishkaoff/ts-backend/internal/domain/lib/utils"
)

// BuildHeaderMap builds map with key: "header name" value: "column index"
func BuildHeaderMap(header []string) map[string]int {
	m := make(map[string]int)
	for i, h := range header {
		m[utils.NormalizeString(h)] = i
	}
	return m
}

// IsEmptyRow checks if all cells in row == ""
func IsEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

// GetCellValueByHeader
// headerMap = map[string]int{
// "price":0
// "name":1
// "partNumber":2
// }
//
// row = []string{"120"(0), "some part"(1), "333"(2)}
//
// finds cell index in `row` header value
func GetCellValueByHeader(
	row []string,
	headerMap map[string]int,
	key string,
) string {
	if collIndex, ok := headerMap[utils.NormalizeString(key)]; ok {
		if collIndex < len(row) {
			return row[collIndex]
		}
	}
	return ""
}
