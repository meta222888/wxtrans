package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	colorAppBg      = color.NRGBA{R: 244, G: 247, B: 249, A: 255}
	colorCardBg     = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	colorCardBorder = color.NRGBA{R: 226, G: 232, B: 240, A: 255}
)

const footerHeight = 22

func newCard(content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(colorCardBg)
	bg.CornerRadius = 10
	bg.StrokeColor = colorCardBorder
	bg.StrokeWidth = 1
	return container.NewStack(bg, container.NewPadded(content))
}

func appBackground(content fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(colorAppBg)
	return container.NewStack(bg, container.NewPadded(content))
}

func formField(label string, field fyne.CanvasObject) fyne.CanvasObject {
	lbl := widget.NewLabel(label)
	lbl.Importance = widget.MediumImportance
	return container.NewVBox(lbl, field)
}

func fieldRow(cols int, fields ...fyne.CanvasObject) fyne.CanvasObject {
	return container.NewGridWithColumns(cols, fields...)
}

// filterRow1：关键词占左侧剩余宽度，右侧并排起始/截止日期（对齐设计图第一行）
func filterRow1(keyword, dateFrom, dateTo fyne.CanvasObject) fyne.CanvasObject {
	datePair := container.NewGridWithColumns(2, dateFrom, dateTo)
	dateWrap := container.New(&fixedWidthLayout{width: 340}, datePair)
	return container.NewBorder(nil, nil, nil, dateWrap, keyword)
}

func newFooterBar(left *widget.Label, right *canvas.Text) fyne.CanvasObject {
	left.Importance = widget.MediumImportance

	row := container.NewBorder(nil, nil, left, right, nil)
	return container.New(&fixedHeightLayout{height: footerHeight}, row)
}

type fixedWidthLayout struct {
	width float32
}

func (l *fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	h := float32(0)
	for _, o := range objects {
		if ms := o.MinSize(); ms.Height > h {
			h = ms.Height
		}
	}
	return fyne.NewSize(l.width, h)
}

func (l *fixedWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	w := l.width
	if size.Width < w {
		w = size.Width
	}
	for _, o := range objects {
		ms := o.MinSize()
		o.Resize(fyne.NewSize(w, max(ms.Height, size.Height)))
		o.Move(fyne.NewPos(0, 0))
	}
}

type fixedHeightLayout struct {
	height float32
}

func (l *fixedHeightLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w := float32(0)
	for _, o := range objects {
		if ms := o.MinSize(); ms.Width > w {
			w = ms.Width
		}
	}
	return fyne.NewSize(w, l.height)
}

func (l *fixedHeightLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	h := l.height
	for _, o := range objects {
		o.Resize(fyne.NewSize(size.Width, h))
		o.Move(fyne.NewPos(0, (h-o.Size().Height)/2))
	}
}
