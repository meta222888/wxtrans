package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"time"

	"wxtrans/internal/database"
	"wxtrans/internal/filedialog"
	"wxtrans/internal/importer"
	"wxtrans/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	window fyne.Window
	db     *database.DB

	keywordEntry    *widget.Entry
	dateFromEntry   *widget.DateEntry
	dateToEntry     *widget.DateEntry
	amountMinEntry  *widget.Entry
	amountMaxEntry  *widget.Entry
	directionSelect *widget.Select
	typeSelect      *widget.Select
	recordCountLabel *widget.Label
	dbPathText       *canvas.Text
	summaryCards     *summaryCards
	table           *widget.Table
	tableData       []models.Transaction
	totalCount      int
	pageSize        int
	pageOffset      int
	pageLabel       *widget.Label
	typeTable       *widget.Table
	typeData        []models.TypeSummary
	monthTable      *widget.Table
	monthData       []models.MonthSummary
}

func NewApp(window fyne.Window, db *database.DB) *App {
	return &App{
		window:   window,
		db:       db,
		pageSize: 100,
	}
}

func (a *App) Build() fyne.CanvasObject {
	a.keywordEntry = widget.NewEntry()
	a.keywordEntry.SetPlaceHolder("对方、商品、单号、备注…")

	a.dateFromEntry = newOptionalDateEntry()
	a.dateFromEntry.SetPlaceHolder("yyyy/mm/dd")
	a.dateToEntry = newOptionalDateEntry()
	a.dateToEntry.SetPlaceHolder("yyyy/mm/dd")

	a.amountMinEntry = widget.NewEntry()
	a.amountMinEntry.SetPlaceHolder("0 = 不限")
	a.amountMaxEntry = widget.NewEntry()
	a.amountMaxEntry.SetPlaceHolder("0 = 不限")

	a.directionSelect = widget.NewSelect([]string{"全部", "收入", "支出"}, func(string) {})
	a.directionSelect.SetSelected("全部")

	a.typeSelect = widget.NewSelect([]string{"全部类型"}, func(string) {})
	a.typeSelect.SetSelected("全部类型")

	searchBtn := widget.NewButtonWithIcon("搜索", theme.SearchIcon(), func() {
		a.pageOffset = 0
		a.refreshList()
		a.refreshSummary()
	})
	searchBtn.Importance = widget.HighImportance

	clearBtn := widget.NewButtonWithIcon("清空条件", theme.ContentClearIcon(), func() { a.clearFilters() })
	importBtn := widget.NewButtonWithIcon("导入 Excel", theme.DocumentCreateIcon(), func() { a.importExcel() })
	passwordBtn := widget.NewButtonWithIcon("改密码", theme.LoginIcon(), func() { a.showChangePasswordDialog() })
	prevBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() { a.prevPage() })
	nextBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() { a.nextPage() })

	a.pageLabel = widget.NewLabel("第 1 / 1 页")
	a.recordCountLabel = widget.NewLabel("共 0 条记录")
	a.dbPathText = canvas.NewText("", color.NRGBA{R: 55, G: 65, B: 81, A: 255})
	a.dbPathText.TextSize = theme.TextSize()

	cards, statCardsRow := newSummaryCards()
	a.summaryCards = cards

	keywordField := formField("关键词", container.NewBorder(
		nil, nil,
		container.NewCenter(widget.NewIcon(theme.SearchIcon())),
		nil,
		a.keywordEntry,
	))

	filterCard := newCard(container.NewVBox(
		filterRow1(
			keywordField,
			formField("起始日期", a.dateFromEntry),
			formField("截止日期", a.dateToEntry),
		),
		fieldRow(4,
			formField("最低金额", a.amountMinEntry),
			formField("最高金额", a.amountMaxEntry),
			formField("收支", a.directionSelect),
			formField("类型", a.typeSelect),
		),
		container.NewBorder(
			nil, nil,
			container.NewHBox(searchBtn, clearBtn, importBtn, passwordBtn),
			container.NewHBox(prevBtn, a.pageLabel, nextBtn),
			nil,
		),
	))

	a.table = widget.NewTable(
		func() (int, int) { return len(a.tableData), len(tableHeaders) },
		func() fyne.CanvasObject { return newTableCellLabel() },
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row >= len(a.tableData) {
				label.SetText("")
				return
			}
			applyTransactionCell(label, id.Col, a.tableData[id.Row])
		},
	)
	headerTable := widget.NewTable(
		func() (int, int) { return 1, len(tableHeaders) },
		func() fyne.CanvasObject { return newTableHeaderLabel("") },
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			if id.Col < len(tableHeaders) {
				obj.(*widget.Label).SetText(tableHeaders[id.Col])
			}
		},
	)

	tablePanel := newFullWidthTablePanel(headerTable, a.table)
	tableCard := newCard(tablePanel)

	footerBar := newFooterBar(a.recordCountLabel, a.dbPathText)

	listTab := container.NewBorder(
		container.NewPadded(filterCard),
		footerBar,
		nil, nil,
		tableCard,
	)

	a.typeTable = widget.NewTable(
		func() (int, int) { return len(a.typeData), 4 },
		func() fyne.CanvasObject { return newTableCellLabel() },
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row >= len(a.typeData) {
				label.SetText("")
				return
			}
			item := a.typeData[id.Row]
			cols := []string{item.TransType, item.Direction, strconv.Itoa(item.Count), database.FormatMoney(item.Amount)}
			label.SetText(cols[id.Col])
		},
	)
	a.typeTable.SetColumnWidth(0, 160)
	a.typeTable.SetColumnWidth(1, 60)
	a.typeTable.SetColumnWidth(2, 60)
	a.typeTable.SetColumnWidth(3, 100)

	typeHeader := widget.NewTable(
		func() (int, int) { return 1, 4 },
		func() fyne.CanvasObject { return newTableHeaderLabel("") },
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			headers := []string{"类型", "收支", "笔数", "金额"}
			if id.Col < len(headers) {
				obj.(*widget.Label).SetText(headers[id.Col])
			}
		},
	)
	typeHeader.SetColumnWidth(0, 160)
	typeHeader.SetColumnWidth(1, 60)
	typeHeader.SetColumnWidth(2, 60)
	typeHeader.SetColumnWidth(3, 100)

	a.monthTable = widget.NewTable(
		func() (int, int) { return len(a.monthData), 4 },
		func() fyne.CanvasObject { return newTableCellLabel() },
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row >= len(a.monthData) {
				label.SetText("")
				return
			}
			item := a.monthData[id.Row]
			cols := []string{
				item.Month,
				database.FormatMoney(item.IncomeAmount),
				database.FormatMoney(item.ExpenseAmount),
				database.FormatMoney(item.NetAmount),
			}
			label.SetText(cols[id.Col])
		},
	)
	a.monthTable.SetColumnWidth(0, 80)
	a.monthTable.SetColumnWidth(1, 100)
	a.monthTable.SetColumnWidth(2, 100)
	a.monthTable.SetColumnWidth(3, 100)

	monthHeader := widget.NewTable(
		func() (int, int) { return 1, 4 },
		func() fyne.CanvasObject { return newTableHeaderLabel("") },
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			headers := []string{"月份", "收入", "支出", "结余"}
			if id.Col < len(headers) {
				obj.(*widget.Label).SetText(headers[id.Col])
			}
		},
	)
	monthHeader.SetColumnWidth(0, 80)
	monthHeader.SetColumnWidth(1, 100)
	monthHeader.SetColumnWidth(2, 100)
	monthHeader.SetColumnWidth(3, 100)

	typeSection := summaryTableSection("按类型汇总", typeHeader, a.typeTable)
	monthSection := summaryTableSection("按月汇总", monthHeader, a.monthTable)
	summarySplit := container.NewHSplit(typeSection, monthSection)
	summarySplit.SetOffset(0.5)

	summaryTab := container.NewBorder(
		container.NewPadded(statCardsRow),
		nil, nil, nil,
		summarySplit,
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("流水", listTab),
		container.NewTabItem("汇总", summaryTab),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	a.refreshTypeOptions()
	a.refreshList()
	a.refreshSummary()

	return appBackground(tabs)
}

func (a *App) currentFilter() models.SearchFilter {
	filter := models.SearchFilter{
		Keyword:   a.keywordEntry.Text,
		Direction: a.directionSelect.Selected,
		Limit:     a.pageSize,
		Offset:    a.pageOffset,
	}
	if t := a.typeSelect.Selected; t != "" && t != "全部类型" {
		filter.TransType = t
	}
	if from := a.dateFromEntry.Date; from != nil {
		t := from.Truncate(24 * time.Hour)
		filter.DateFrom = &t
	}
	if to := a.dateToEntry.Date; to != nil {
		t := to.Truncate(24 * time.Hour)
		filter.DateTo = &t
	}
	filter.AmountMin = parseAmountFilter(a.amountMinEntry.Text)
	filter.AmountMax = parseAmountFilter(a.amountMaxEntry.Text)
	return filter
}

func parseAmountFilter(raw string) float64 {
	raw = stringsTrim(raw)
	if raw == "" || raw == "0" {
		return 0
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v <= 0 {
		return 0
	}
	return v
}

func (a *App) refreshList() {
	filter := a.currentFilter()
	list, total, err := a.db.Search(filter)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	a.tableData = list
	a.totalCount = total
	a.table.Refresh()

	page := a.pageOffset/a.pageSize + 1
	totalPages := max(1, (total+a.pageSize-1)/a.pageSize)
	a.pageLabel.SetText(fmt.Sprintf("第 %d / %d 页", page, totalPages))
	end := min(a.pageOffset+len(list), total)
	if total == 0 {
		a.recordCountLabel.SetText("共 0 条记录")
	} else {
		a.recordCountLabel.SetText(fmt.Sprintf("共 %d 条记录，当前显示 %d-%d 条", total, a.pageOffset+1, end))
	}
	a.dbPathText.Text = "数据库: " + a.db.Path()
	a.dbPathText.Refresh()
}

func (a *App) refreshSummary() {
	filter := a.currentFilter()
	filter.Limit = 0
	filter.Offset = 0

	summary, err := a.db.Summary(filter)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	a.summaryCards.Update(summary)

	typeData, err := a.db.SummaryByType(filter)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	a.typeData = typeData
	a.typeTable.Refresh()

	monthData, err := a.db.SummaryByMonth(filter)
	if err != nil {
		dialog.ShowError(err, a.window)
		return
	}
	a.monthData = monthData
	a.monthTable.Refresh()
}

func (a *App) refreshTypeOptions() {
	types, err := a.db.DistinctTypes()
	if err != nil {
		return
	}
	options := []string{"全部类型"}
	options = append(options, types...)
	a.typeSelect.Options = options
	a.typeSelect.Refresh()
}

func (a *App) clearFilters() {
	a.keywordEntry.SetText("")
	a.dateFromEntry.SetDate(nil)
	a.dateToEntry.SetDate(nil)
	a.amountMinEntry.SetText("")
	a.amountMaxEntry.SetText("")
	a.directionSelect.SetSelected("全部")
	a.typeSelect.SetSelected("全部类型")
	a.pageOffset = 0
	a.refreshList()
	a.refreshSummary()
}

func (a *App) prevPage() {
	if a.pageOffset >= a.pageSize {
		a.pageOffset -= a.pageSize
		a.refreshList()
	}
}

func (a *App) nextPage() {
	if a.pageOffset+a.pageSize < a.totalCount {
		a.pageOffset += a.pageSize
		a.refreshList()
	}
}

func (a *App) importExcel() {
	go func() {
		path, err := filedialog.OpenExcel()
		fyne.Do(func() {
			if err != nil {
				dialog.ShowError(err, a.window)
				return
			}
			if path == "" {
				return
			}
			result, err := importer.ImportWeChatExcel(a.db, path)
			if err != nil {
				dialog.ShowError(err, a.window)
				return
			}
			a.pageOffset = 0
			a.refreshTypeOptions()
			a.refreshList()
			a.refreshSummary()
			dialog.ShowInformation("导入结果", result.Message, a.window)
		})
	}()
}

func stringsTrim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
