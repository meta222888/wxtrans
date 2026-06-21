package ui

import (
	"fmt"
	"strconv"
	"time"

	"wxtrans/internal/database"
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

	keywordEntry   *widget.Entry
	dateFromEntry  *widget.Entry
	dateToEntry    *widget.Entry
	directionSelect *widget.Select
	typeSelect     *widget.Select
	statusLabel    *widget.Label
	summaryLabel   *widget.Label
	table          *widget.Table
	tableData      []models.Transaction
	totalCount     int
	pageSize       int
	pageOffset     int
	pageLabel      *widget.Label
	typeTable      *widget.Table
	typeData       []models.TypeSummary
	monthTable     *widget.Table
	monthData      []models.MonthSummary
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
	prevBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() { a.prevPage() })
	nextBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() { a.nextPage() })

	a.pageLabel = widget.NewLabel("第 1 页")
	a.statusLabel = widget.NewLabel("就绪")
	a.summaryLabel = widget.NewLabel("汇总加载中…")

	filterBar := container.NewGridWithColumns(4,
		widget.NewForm(
			widget.NewFormItem("关键词", a.keywordEntry),
			widget.NewFormItem("起始日期", a.dateFromEntry),
			widget.NewFormItem("截止日期", a.dateToEntry),
		),
		widget.NewForm(
			widget.NewFormItem("收支", a.directionSelect),
			widget.NewFormItem("类型", a.typeSelect),
		),
		container.NewHBox(searchBtn, clearBtn, importBtn),
		container.NewHBox(prevBtn, a.pageLabel, nextBtn),
	)

	a.table = widget.NewTable(
		func() (int, int) { return len(a.tableData), 8 },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
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
	a.table.SetColumnWidth(0, 130)
	a.table.SetColumnWidth(1, 90)
	a.table.SetColumnWidth(2, 120)
	a.table.SetColumnWidth(3, 50)
	a.table.SetColumnWidth(4, 80)
	a.table.SetColumnWidth(5, 130)
	a.table.SetColumnWidth(6, 80)
	a.table.SetColumnWidth(7, 180)

	header := container.NewGridWithColumns(8,
		widget.NewLabelWithStyle("时间", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("类型", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("对方", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("收支", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("金额", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("支付方式", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("状态", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("交易单号", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	listTab := container.NewBorder(
		filterBar,
		a.statusLabel,
		nil, nil,
		container.NewBorder(header, nil, nil, nil, a.table),
	)

	a.typeTable = widget.NewTable(
		func() (int, int) { return len(a.typeData), 4 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
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
	a.typeTable.SetColumnWidth(0, 140)
	a.typeTable.SetColumnWidth(1, 60)
	a.typeTable.SetColumnWidth(2, 60)
	a.typeTable.SetColumnWidth(3, 100)

	a.monthTable = widget.NewTable(
		func() (int, int) { return len(a.monthData), 4 },
		func() fyne.CanvasObject { return widget.NewLabel("") },
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

	summaryTab := container.NewVBox(
		a.summaryLabel,
		widget.NewLabelWithStyle("按类型汇总", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(4,
			widget.NewLabelWithStyle("类型", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("收支", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("笔数", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("金额", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		),
		container.NewMax(a.typeTable),
		widget.NewLabelWithStyle("按月汇总", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewGridWithColumns(4,
			widget.NewLabelWithStyle("月份", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("收入", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("支出", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("结余", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		),
		container.NewMax(a.monthTable),
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
	return filter
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
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()
		path := reader.URI().Path()
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
	}, a.window)
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
