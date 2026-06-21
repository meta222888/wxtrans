package ui

import (
	"strings"

	"wxtrans/internal/database"
	"wxtrans/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var tableHeaders = []string{"时间", "类型", "对方", "收支", "金额", "支付方式", "状态", "交易单号"}

var tableColWidths = []float32{132, 88, 148, 76, 88, 118, 96, 148}

func newTableCellLabel() *widget.Label {
	l := widget.NewLabel("")
	l.Truncation = fyne.TextTruncateEllipsis
	return l
}

func newTableHeaderLabel(text string) *widget.Label {
	l := widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	l.Truncation = fyne.TextTruncateEllipsis
	l.Importance = widget.MediumImportance
	return l
}

func applyTableColumnWidths(table *widget.Table) {
	for i, w := range tableColWidths {
		table.SetColumnWidth(i, w)
	}
}

func applyTransactionCell(label *widget.Label, col int, tx models.Transaction) {
	label.Importance = widget.MediumImportance
	label.TextStyle = fyne.TextStyle{}

	switch col {
	case 0:
		label.SetText(tx.TransTime.Format("2006-01-02 15:04"))
		label.Importance = widget.LowImportance
	case 1:
		label.SetText(tx.TransType)
	case 2:
		label.SetText(tx.Counterparty)
	case 3:
		label.SetText(formatDirectionBadge(tx.Direction))
		switch strings.TrimSpace(tx.Direction) {
		case "收入":
			label.Importance = widget.SuccessImportance
		case "支出":
			label.Importance = widget.DangerImportance
		default:
			label.Importance = widget.MediumImportance
		}
	case 4:
		label.SetText(formatAmount(tx.Direction, tx.Amount))
		switch strings.TrimSpace(tx.Direction) {
		case "收入":
			label.Importance = widget.SuccessImportance
		case "支出":
			label.Importance = widget.DangerImportance
		default:
			label.Importance = widget.MediumImportance
		}
	case 5:
		label.SetText(tx.PaymentMethod)
	case 6:
		label.SetText(formatStatus(tx.Status))
		if isSuccessStatus(tx.Status) {
			label.Importance = widget.SuccessImportance
		} else {
			label.Importance = widget.MediumImportance
		}
	case 7:
		label.SetText(tx.TransNo)
	default:
		label.SetText("")
	}
}

func formatDirectionBadge(direction string) string {
	switch strings.TrimSpace(direction) {
	case "收入":
		return "↑ 收入"
	case "支出":
		return "↓ 支出"
	default:
		return "—"
	}
}

func formatAmount(direction string, amount float64) string {
	money := database.FormatMoney(amount)
	switch strings.TrimSpace(direction) {
	case "收入":
		return "+ ¥" + money
	case "支出":
		return "- ¥" + money
	default:
		if amount > 0 {
			return "¥" + money
		}
		return money
	}
}

func isSuccessStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "已收钱", "支付成功", "已存入零钱", "提现已到账", "已转账", "已领取":
		return true
	default:
		return false
	}
}

func formatStatus(status string) string {
	if status == "" {
		return ""
	}
	return "● " + status
}

func summaryTableSection(title string, header, table *widget.Table) fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	top := container.NewVBox(titleLabel, header)
	return container.NewBorder(top, nil, nil, nil, container.NewScroll(table))
}
