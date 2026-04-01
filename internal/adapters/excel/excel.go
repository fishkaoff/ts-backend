package excel_adapter

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"

	"github.com/fishkaoff/ts-backend/internal/domain/types"
	"github.com/xuri/excelize/v2"
)

var ErrEmptyFile = errors.New("Файл пустой")
var ErrNoSheets = errors.New("Не удалось найти листы в файле")

type ExcelAdapter struct {
}

func New() *ExcelAdapter {
	return &ExcelAdapter{}
}

// ParseProductsFromFile parses products from excel file
func (a *ExcelAdapter) ParseProductsFromFile(ctx context.Context, file multipart.File) ([]types.Product, error) {
	const op = "adapters.excel.parseProducts"

	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to open excel: %w", op, err)
	}

	sheet := f.GetSheetName(0)
	if sheet == "" {
		return nil, ErrNoSheets
	}

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read rows: %w", op, err)
	}

	if len(rows) == 0 {
		return nil, ErrEmptyFile
	}

	headerMap := BuildHeaderMap(rows[0])
	products := make([]types.Product, 0, len(rows)-1)

	for _, row := range rows[1:] {
		if IsEmptyRow(row) {
			continue
		}

		code, _ := strconv.Atoi(
			GetCellValueByHeader(row, headerMap, "Код"),
		)

		balance, err := strconv.ParseInt(GetCellValueByHeader(row, headerMap, "Остаток"), 10, 64)
		if err != nil {
			balance = -1
		}

		product := types.Product{
			Code:         code,
			PartNumber:   GetCellValueByHeader(row, headerMap, "Артикул"),
			Name:         GetCellValueByHeader(row, headerMap, "Наименование"),
			Manufacturer: GetCellValueByHeader(row, headerMap, "Производитель"),
			Unit:         types.ParseUnit(GetCellValueByHeader(row, headerMap, "Ед")),
			Price:        types.ParsePrice(GetCellValueByHeader(row, headerMap, "Цена,руб")),
			IsNew:        types.ParseIsNew(GetCellValueByHeader(row, headerMap, "Новинка!")),
			ImageURL:     GetCellValueByHeader(row, headerMap, "Изображение"),
			Balance:      balance,
			Active:       true,
		}

		if product.Name == "" {
			continue
		}

		products = append(products, product)
	}

	return products, nil
}
