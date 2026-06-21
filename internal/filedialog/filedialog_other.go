//go:build !windows

package filedialog

import "errors"

func OpenExcel() (string, error) {
	return "", errors.New("系统文件选择框仅支持 Windows")
}
