package filedialog

import (
	"errors"

 dlg "github.com/sqweek/dialog"
)

// OpenExcel 打开系统原生文件选择框，返回 Excel 路径；取消时返回空字符串。
func OpenExcel() (string, error) {
	path, err := dlg.File().
		Filter("Excel 文件", "xlsx", "xls").
		Title("选择微信支付账单").
		Load()
	if err != nil {
		if errors.Is(err, dlg.ErrCancelled) {
			return "", nil
		}
		return "", err
	}
	return path, nil
}
