package ui

import (
	"fmt"
	"strings"
	"time"

	"wxtrans/internal/database"
	"wxtrans/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var tableHeaders = []string{"时间", "类型", "对方", "收支", "金额", "支付方式", "状态", "交易单号"}

// 列宽比例，合计 1.0
var tableColRatios = []float32{0.11, 0.08, 0.14, 0.06, 0.08, 0.11, 0.09, 0.33}

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

func applyTableColumnWidths(table *widget.Table, totalWidth float32) {
	if totalWidth <= 0 {
		return
	}
	for i, ratio := range tableColRatios {
		table.SetColumnWidth(i, totalWidth*ratio)
	}
}

func newFullWidthTablePanel(header, body *widget.Table) fyne.CanvasObject {
	scroll := container.NewScroll(body)
	return container.New(&fullWidthTableLayout{
		header: header,
		body:   body,
		scroll: scroll,
	}, header, scroll)
}

type fullWidthTableLayout struct {
	header *widget.Table
	body   *widget.Table
	scroll *container.Scroll
}

func (l *fullWidthTableLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	hh := l.header.MinSize().Height
	sh := l.scroll.MinSize().Height
	sw := float32(0)
	for _, o := range objects {
		if ms := o.MinSize(); ms.Width > sw {
			sw = ms.Width
		}
	}
	return fyne.NewSize(sw, hh+sh)
}

func (l *fullWidthTableLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	w := size.Width
	applyTableColumnWidths(l.header, w)
	applyTableColumnWidths(l.body, w)

	hh := l.header.MinSize().Height
	l.header.Move(fyne.NewPos(0, 0))
	l.header.Resize(fyne.NewSize(w, hh))

	l.scroll.Move(fyne.NewPos(0, hh))
	l.scroll.Resize(fyne.NewSize(w, size.Height-hh))
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

func newOptionalDateEntry() *widget.DateEntry {
	e := widget.NewDateEntry()
	_ = e.MinSize()
	e.Validator = optionalDateValidator
	return e
}

func optionalDateValidator(in string) error {
	if strings.TrimSpace(in) == "" {
		return nil
	}
	for _, layout := range []string{"2006/01/02", "2006-01-02", "2006/1/2", "2006-1-2"} {
		if _, err := time.Parse(layout, in); err == nil {
			return nil
		}
	}
	return fmt.Errorf("日期格式无效")
}
