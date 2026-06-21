package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var tableHeaders = []string{"时间", "类型", "对方", "收支", "金额", "支付方式", "状态", "交易单号"}

var tableColWidths = []float32{132, 88, 148, 44, 72, 118, 72, 148}

func newTableCellLabel() *widget.Label {
	l := widget.NewLabel("")
	l.Truncation = fyne.TextTruncateEllipsis
	return l
}

func newTableHeaderLabel(text string) *widget.Label {
	l := widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	l.Truncation = fyne.TextTruncateEllipsis
	return l
}

func applyTableColumnWidths(table *widget.Table) {
	for i, w := range tableColWidths {
		table.SetColumnWidth(i, w)
	}
}
