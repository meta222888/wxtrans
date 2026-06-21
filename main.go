package main

import (
	"flag"
	"log"

	"wxtrans/internal/database"
	"wxtrans/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	dbPath := flag.String("db", "", "SQLite 数据库路径（默认 %AppData%/wxtrans/ledger.db）")
	flag.Parse()

	db, err := database.Open(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	application := app.NewWithID("com.wxtrans.ledger")
	window := application.NewWindow("微信小账本")
	window.Resize(fyne.NewSize(380, 180))

	ui.ShowLogin(window, db, func() {
		window.SetTitle("微信小账本")
		window.Resize(fyne.NewSize(980, 640))
		window.CenterOnScreen()
		ledger := ui.NewApp(window, db)
		window.SetContent(ledger.Build())
	})

	window.ShowAndRun()
}
