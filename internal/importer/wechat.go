package importer

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"wxtrans/internal/database"
	"wxtrans/internal/models"

	"github.com/xuri/excelize/v2"
)

const headerMarker = "交易单号"

var headerColumns = map[string]string{
	"交易时间":  "trans_time",
	"交易类型":  "trans_type",
	"交易对方":  "counterparty",
	"商品":    "product",
	"收/支":   "direction",
	"金额(元)": "amount",
	"支付方式":  "payment_method",
	"当前状态":  "status",
	"交易单号":  "trans_no",
	"商户单号":  "merchant_no",
	"备注":    "remark",
}

func ImportWeChatExcel(db *database.DB, filePath string) (*models.ImportResult, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开 Excel 失败: %w", err)
	}
	defer f.Close()

	sheet := f.GetSheetName(0)
	if sheet == "" {
		return nil, fmt.Errorf("Excel 中没有工作表")
	}

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("读取 Excel 失败: %w", err)
	}

	headerRow, colIndex, err := findHeader(rows)
	if err != nil {
		return nil, err
	}

	result := &models.ImportResult{}
	for i := headerRow + 1; i < len(rows); i++ {
		row := rows[i]
		if isEmptyRow(row) {
			continue
		}
		tx, err := parseRow(row, colIndex)
		if err != nil {
			result.Errors++
			continue
		}
		if tx.TransNo == "" {
			result.Errors++
			continue
		}
		inserted, err := db.InsertTransaction(tx)
		if err != nil {
			return nil, fmt.Errorf("写入数据库失败: %w", err)
		}
		if inserted {
			result.Imported++
		} else {
			result.Skipped++
		}
	}

	result.Message = fmt.Sprintf("导入完成：新增 %d 条，跳过重复 %d 条，无效 %d 条",
		result.Imported, result.Skipped, result.Errors)
	return result, nil
}

func findHeader(rows [][]string) (int, map[string]int, error) {
	for i, row := range rows {
		idx := map[string]int{}
		for j, cell := range row {
			name := strings.TrimSpace(cell)
			if _, ok := headerColumns[name]; ok {
				idx[name] = j
			}
		}
		if _, ok := idx[headerMarker]; ok {
			return i, idx, nil
		}
	}
	return -1, nil, fmt.Errorf("未找到微信支付账单表头（需包含「交易单号」列）")
}

func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func parseRow(row []string, colIndex map[string]int) (*models.Transaction, error) {
	get := func(name string) string {
		i, ok := colIndex[name]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	transTime, err := parseTime(get("交易时间"))
	if err != nil {
		return nil, err
	}
	amount, err := parseAmount(get("金额(元)"))
	if err != nil {
		return nil, err
	}

	return &models.Transaction{
		TransTime:     transTime,
		TransType:     get("交易类型"),
		Counterparty:  get("交易对方"),
		Product:       get("商品"),
		Direction:     get("收/支"),
		Amount:        amount,
		PaymentMethod: get("支付方式"),
		Status:        get("当前状态"),
		TransNo:       get("交易单号"),
		MerchantNo:    get("商户单号"),
		Remark:        get("备注"),
	}, nil
}

func parseTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, fmt.Errorf("empty time")
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"2006-01-02 15:04",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return t, nil
		}
	}
	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		t, err := excelize.ExcelDateToTime(f, false)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time: %s", raw)
}

func parseAmount(raw string) (float64, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "¥")
	raw = strings.TrimPrefix(raw, "￥")
	raw = strings.ReplaceAll(raw, ",", "")
	if raw == "" || raw == "/" {
		return 0, nil
	}
	return strconv.ParseFloat(raw, 64)
}
