package ui

import (
	"wxtrans/internal/database"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func ShowLogin(window fyne.Window, db *database.DB, onSuccess func()) {
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("请输入密码")

	loginBtn := widget.NewButton("登录", nil)
	loginBtn.Importance = widget.HighImportance

	tryLogin := func() {
		if err := db.VerifyPassword(passwordEntry.Text); err != nil {
			dialog.ShowError(err, window)
			passwordEntry.SetText("")
			passwordEntry.FocusGained()
			return
		}
		onSuccess()
	}

	loginBtn.OnTapped = tryLogin
	passwordEntry.OnSubmitted = func(string) { tryLogin() }

	content := container.NewVBox(
		widget.NewLabel("请输入密码以打开账本"),
		passwordEntry,
		loginBtn,
	)

	window.SetContent(container.NewPadded(content))
	window.Resize(fyne.NewSize(380, 150))
	window.CenterOnScreen()
	passwordEntry.FocusGained()
}

func (a *App) showChangePasswordDialog() {
	oldEntry := widget.NewPasswordEntry()
	newEntry := widget.NewPasswordEntry()
	confirmEntry := widget.NewPasswordEntry()

	form := widget.NewForm(
		widget.NewFormItem("当前密码", oldEntry),
		widget.NewFormItem("新密码", newEntry),
		widget.NewFormItem("确认新密码", confirmEntry),
	)

	d := dialog.NewCustomConfirm("修改密码", "保存", "取消", form, func(ok bool) {
		if !ok {
			return
		}
		if err := a.db.ChangePassword(oldEntry.Text, newEntry.Text, confirmEntry.Text); err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		dialog.ShowInformation("成功", "密码已修改", a.window)
	}, a.window)
	d.Resize(fyne.NewSize(360, 220))
	d.Show()
}
