package cursesgraphic

import (
	"image/color"
)

type vector struct {
	X int
	Y int
}

type canContext struct {
	fillChar          rune
	textColor         color.RGBA
	backgroundColor   color.RGBA
	translationVector vector
}

type Canvas struct {
	width      int
	height     int
	saveStates []canContext
	//used to show where a line may have its colour broken up
	lineBreakage []int
	contents     [][]runeCell
	graphicLines []lineInfo
}

type lineInfo struct {
	x1, y1, x2, y2 int
}

type runeCell struct {
	character       rune
	textColor       color.RGBA
	backgroundColor color.RGBA
}
