package ui

import (
	"fmt"
	"image/color"

	"wxtrans/internal/database"
	"wxtrans/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	colorIncome       = color.NRGBA{R: 22, G: 163, B: 74, A: 255}
	colorIncomeBadge  = color.NRGBA{R: 220, G: 252, B: 231, A: 255}
	colorExpense      = color.NRGBA{R: 220, G: 38, B: 38, A: 255}
	colorExpenseBadge = color.NRGBA{R: 254, G: 226, B: 226, A: 255}
	colorNetGradStart = color.NRGBA{R: 34, G: 197, B: 94, A: 255}
	colorSubText      = color.NRGBA{R: 100, G: 116, B: 139, A: 255}
	colorWhite        = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
)

const summaryCardHeight = 108

type summaryCards struct {
	incomeAmount  *canvas.Text
	incomeSub     *canvas.Text
	expenseAmount *canvas.Text
	expenseSub    *canvas.Text
	netAmount     *canvas.Text
	netSub        *canvas.Text
}

func newSummaryCards() (*summaryCards, fyne.CanvasObject) {
	c := &summaryCards{}

	c.incomeAmount = statAmountText(colorIncome)
	c.incomeSub = statSubText(colorSubText)
	c.expenseAmount = statAmountText(colorExpense)
	c.expenseSub = statSubText(colorSubText)
	c.netAmount = statAmountText(colorWhite)
	c.netSub = statSubText(colorWhite)

	incomeCard := newWhiteStatCard("本期收入", c.incomeAmount, c.incomeSub, theme.MoveUpIcon(), colorIncomeBadge)
	expenseCard := newWhiteStatCard("本期支出", c.expenseAmount, c.expenseSub, theme.MoveDownIcon(), colorExpenseBadge)
	netCard := newNetStatCard(c.netAmount, c.netSub)

	row := container.NewGridWithColumns(3, incomeCard, expenseCard, netCard)
	return c, row
}

func (c *summaryCards) Update(s *models.Summary) {
	c.incomeAmount.Text = "+ ¥" + database.FormatMoney(s.IncomeAmount)
	c.expenseAmount.Text = "- ¥" + database.FormatMoney(s.ExpenseAmount)
	c.netAmount.Text = "¥" + database.FormatMoney(s.NetAmount)
	c.incomeSub.Text = fmt.Sprintf("%d 笔到账", s.IncomeCount)
	c.expenseSub.Text = fmt.Sprintf("%d 笔消费", s.ExpenseCount)
	c.netSub.Text = fmt.Sprintf("%d 笔交易", s.TotalCount)

	c.incomeAmount.Refresh()
	c.expenseAmount.Refresh()
	c.netAmount.Refresh()
	c.incomeSub.Refresh()
	c.expenseSub.Refresh()
	c.netSub.Refresh()
}

func statAmountText(c color.Color) *canvas.Text {
	t := canvas.NewText("¥0.00", c)
	t.TextSize = 22
	t.TextStyle = fyne.TextStyle{Bold: true}
	return t
}

func statSubText(c color.Color) *canvas.Text {
	t := canvas.NewText("", c)
	t.TextSize = theme.TextSize() - 1
	return t
}

func statSubTextWhite() *canvas.Text {
	t := statSubText(colorWhite)
	t.TextSize = theme.TextSize() - 1
	return t
}

func statTitleText(text string, c color.Color) *canvas.Text {
	t := canvas.NewText(text, c)
	t.TextSize = theme.TextSize()
	return t
}

func iconBadge(icon fyne.Resource, bg color.Color) fyne.CanvasObject {
	bgRect := canvas.NewRectangle(bg)
	bgRect.CornerRadius = 14
	ic := widget.NewIcon(icon)
	return container.NewStack(
		container.NewGridWrap(fyne.NewSize(28, 28), bgRect),
		container.NewCenter(ic),
	)
}

func newWhiteStatCard(title string, amount, sub *canvas.Text, icon fyne.Resource, badgeBg color.Color) fyne.CanvasObject {
	titleText := statTitleText(title, colorSubText)
	top := container.NewBorder(nil, nil, titleText, iconBadge(icon, badgeBg), nil)
	content := container.NewVBox(
		top,
		amount,
		sub,
	)
	card := newCard(content)
	return container.New(&minHeightLayout{height: summaryCardHeight}, card)
}

func newNetStatCard(amount, sub *canvas.Text) fyne.CanvasObject {
	bg := canvas.NewRectangle(colorNetGradStart)
	bg.CornerRadius = 10

	title := statTitleText("净结余", colorWhite)
	badge := iconBadge(theme.DocumentIcon(), color.NRGBA{R: 255, G: 255, B: 255, A: 60})
	top := container.NewBorder(nil, nil, title, badge, nil)
	content := container.NewPadded(container.NewVBox(top, amount, sub))
	inner := container.NewStack(bg, content)
	return container.New(&minHeightLayout{height: summaryCardHeight}, inner)
}

type minHeightLayout struct {
	height float32
}

func (l *minHeightLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w := float32(0)
	h := l.height
	for _, o := range objects {
		ms := o.MinSize()
		if ms.Width > w {
			w = ms.Width
		}
		if ms.Height > h {
			h = ms.Height
		}
	}
	return fyne.NewSize(w, h)
}

func (l *minHeightLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	h := max(l.height, size.Height)
	for _, o := range objects {
		o.Resize(fyne.NewSize(size.Width, h))
		o.Move(fyne.NewPos(0, 0))
	}
}

func summaryTableSection(title string, header, table *widget.Table) fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	body := container.NewBorder(
		container.NewVBox(titleLabel, header),
		nil, nil, nil,
		container.NewScroll(table),
	)
	return newCard(body)
}
