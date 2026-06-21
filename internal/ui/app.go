package ui

import (
	"fmt"
	"strconv"
	"time"

	"wxtrans/internal/database"
	"wxtrans/internal/filedialog"
	"wxtrans/internal/importer"
	"wxtrans/internal/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	window fyne.Window
	db     *database.DB

	keywordEntry    *widget.Entry
	dateFromEntry   *widget.Entry
	dateToEntry     *widget.Entry
	amountMinEntry  *widget.Entry
	amountMaxEntry  *widget.Entry
	directionSelect *widget.Select
	typeSelect      *widget.Select
	statusLabel     *widget.Label
	summaryLabel    *widget.Label
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
	a.keywordEntry.SetPlaceHolder("关键词：对方、商品、单号、备注…")

	a.dateFromEntry = widget.NewEntry()
	a.dateFromEntry.SetPlaceHolder("起始 yyyy-mm-dd")
	a.dateToEntry = widget.NewEntry()
	a.dateToEntry.SetPlaceHolder("截止 yyyy-mm-dd")

	a.amountMinEntry = widget.NewEntry()
	a.amountMinEntry.SetPlaceHolder("最低金额，0=不限")
	a.amountMaxEntry = widget.NewEntry()
	a.amountMaxEntry.SetPlaceHolder("最高金额，0=不限")

	a.directionSelect = widget.NewSelect([]string{"全部", "收入", "支出"}, func(string) {})
	a.directionSelect.SetSelected("全部")

	a.typeSelect = widget.NewSelect([]string{"全部类型"}, func(string) {})
	a.typeSelect.SetSelected("全部类型")

	searchBtn := widget.NewButtonWithIcon("搜索", theme.SearchIcon(), func() {
		a.pageOffset = 0
		a.refreshList()
		a.refreshSummary()
	})
	clearBtn := widget.NewButton("清空条件", func() { a.clearFilters() })
	importBtn := widget.NewButtonWithIcon("导入 Excel", theme.DocumentCreateIcon(), func() { a.importExcel() })
	passwordBtn := widget.NewButtonWithIcon("改密码", theme.LoginIcon(), func() { a.showChangePasswordDialog() })
	prevBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() { a.prevPage() })
	nextBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() { a.nextPage() })

	a.pageLabel = widget.NewLabel("第 1 页")
	a.statusLabel = widget.NewLabel("就绪")
	a.summaryLabel = widget.NewLabel("汇总加载中…")

	filterBar := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("关键词", a.keywordEntry),
			widget.NewFormItem("起始日期", a.dateFromEntry),
			widget.NewFormItem("截止日期", a.dateToEntry),
			widget.NewFormItem("最低金额", a.amountMinEntry),
			widget.NewFormItem("最高金额", a.amountMaxEntry),
			widget.NewFormItem("收支", a.directionSelect),
			widget.NewFormItem("类型", a.typeSelect),
		),
		container.NewHBox(searchBtn, clearBtn, importBtn, passwordBtn, prevBtn, a.pageLabel, nextBtn),
	)

	a.table = widget.NewTable(
		func() (int, int) { return len(a.tableData), len(tableHeaders) },
		func() fyne.CanvasObject { return newTableCellLabel() },
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id.Row >= len(a.tableData) {
				label.SetText("")
				return
			}
			tx := a.tableData[id.Row]
			cols := []string{
				tx.TransTime.Format("2006-01-02 15:04"),
				tx.TransType,
				tx.Counterparty,
				tx.Direction,
				database.FormatMoney(tx.Amount),
				tx.PaymentMethod,
				tx.Status,
				tx.TransNo,
			}
			label.SetText(cols[id.Col])
		},
	)
	applyTableColumnWidths(a.table)

	headerTable := widget.NewTable(
		func() (int, int) { return 1, len(tableHeaders) },
		func() fyne.CanvasObject { return newTableHeaderLabel("") },
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			if id.Col < len(tableHeaders) {
				obj.(*widget.Label).SetText(tableHeaders[id.Col])
			}
		},
	)
	applyTableColumnWidths(headerTable)

	listTab := container.NewBorder(
		filterBar,
		a.statusLabel,
		nil, nil,
		container.NewBorder(headerTable, nil, nil, nil, container.NewScroll(a.table)),
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

	summaryTab := container.NewVBox(
		a.summaryLabel,
		widget.NewLabelWithStyle("按类型汇总", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(typeHeader, nil, nil, nil, container.NewScroll(a.typeTable)),
		widget.NewLabelWithStyle("按月汇总", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(monthHeader, nil, nil, nil, container.NewScroll(a.monthTable)),
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("流水", listTab),
		container.NewTabItem("汇总", summaryTab),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	a.refreshTypeOptions()
	a.refreshList()
	a.refreshSummary()

	return tabs
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
	if from := stringsTrim(a.dateFromEntry.Text); from != "" {
		if t, err := time.ParseInLocation("2006-01-02", from, time.Local); err == nil {
			filter.DateFrom = &t
		}
	}
	if to := stringsTrim(a.dateToEntry.Text); to != "" {
		if t, err := time.ParseInLocation("2006-01-02", to, time.Local); err == nil {
			filter.DateTo = &t
		}
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
	a.statusLabel.SetText(fmt.Sprintf("共 %d 条记录，当前显示 %d-%d 条 | 数据库: %s",
		total, a.pageOffset+1, min(a.pageOffset+len(list), total), a.db.Path()))
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
	a.summaryLabel.SetText(fmt.Sprintf(
		"总笔数: %d  |  收入: %d 笔 / ¥%s  |  支出: %d 笔 / ¥%s  |  结余: ¥%s",
		summary.TotalCount,
		summary.IncomeCount, database.FormatMoney(summary.IncomeAmount),
		summary.ExpenseCount, database.FormatMoney(summary.ExpenseAmount),
		database.FormatMoney(summary.NetAmount),
	))

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
	a.dateFromEntry.SetText("")
	a.dateToEntry.SetText("")
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
