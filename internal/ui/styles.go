package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	colorAppBg     = color.NRGBA{R: 244, G: 247, B: 249, A: 255}
	colorCardBg    = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	colorCardBorder = color.NRGBA{R: 226, G: 232, B: 240, A: 255}
)

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
