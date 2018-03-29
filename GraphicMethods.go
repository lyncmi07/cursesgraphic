package cursesgraphic

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten"
)

func init() {
	enterKeyDown = false
}

func (c Canvas) GetRune(x, y int) rune {
	return c.contents[y][x].character
}

func (c Canvas) Width() int {
	return c.width
}

func (c Canvas) Height() int {
	return c.height
}

var enterKeyDown bool

func (c Canvas) GetCharCode() rune {
	currentInput := ebiten.InputChars()
	for len(currentInput) < 1 {
		currentInput = ebiten.InputChars()
		if ebiten.IsKeyPressed(ebiten.KeyEnter) && !enterKeyDown {
			enterKeyDown = true
			return '\n'
		}

		if !ebiten.IsKeyPressed(ebiten.KeyEnter) && enterKeyDown {
			enterKeyDown = false
		}
		time.Sleep(1)
	}
	return currentInput[0]
}

func (c *Canvas) setLineBreakage(lineIndex int) {
	//NOTE:maybe in order? This allows for binary search
	for i := 0; i < len(c.lineBreakage); i++ {
		if c.lineBreakage[i] == lineIndex {
			return
		}
	}

	c.lineBreakage = append(c.lineBreakage, lineIndex)
}

func (c *Canvas) unsetLineBreakage(lineIndex int) {
	for i := 0; i < len(c.lineBreakage); i++ {
		if c.lineBreakage[i] == lineIndex {
			c.lineBreakage = append(c.lineBreakage[:i], c.lineBreakage[(i+1):]...)
		}
	}
}

func (c *Canvas) hasLineBreakage(lineIndex int) bool {
	for i := 0; i < len(c.lineBreakage); i++ {
		if c.lineBreakage[i] == lineIndex {
			return true
		}
	}

	return false
}

func (c *Canvas) Save() {
	newStack := make([]canContext, len(c.saveStates)+1)
	copy(newStack[1:len(newStack)], c.saveStates)
	c.saveStates = newStack
	c.saveStates[0].fillChar = c.saveStates[1].fillChar
	c.saveStates[0].translationVector = c.saveStates[1].translationVector
}

func (c *Canvas) Restore() {
	c.saveStates = c.saveStates[1:len(c.saveStates)]
}

func (c *Canvas) SetColor(background, text color.RGBA) {
	c.saveStates[0].textColor = text
	c.saveStates[0].backgroundColor = background
}

func (c *Canvas) SetTranslate(x int, y int) {
	c.saveStates[0].translationVector.X = x
	c.saveStates[0].translationVector.Y = y
}

func (c *Canvas) Translate(x int, y int) {
	c.saveStates[0].translationVector.X += x
	c.saveStates[0].translationVector.Y += y
}

func (c *Canvas) SetFillChar(ch rune) {
	c.saveStates[0].fillChar = ch
}

//draw a line from between two characters
func (c *Canvas) DrawLine(x1, y1, x2, y2 int) {
	c.graphicLines = append(c.graphicLines, lineInfo{x1, y1, x2, y2})
}

//clears previously drawn lines from view
func (c *Canvas) ClearLines() {
	c.graphicLines = make([]lineInfo, 0)
}

func (c *Canvas) FillText(text string, x, y int) {
	//use this
	x += c.saveStates[0].translationVector.X
	y += c.saveStates[0].translationVector.Y

	aX := x
	aY := y

	//text is off the canvas, do not attempt to write
	if (aY < 0) || (aY >= c.height) {
		return
	}

	//first part of text if of screen, do not attempt to write
	if aX < 0 {
		text = text[(-aX):]
		aX = 0
	}

	for i := aX; (i < (aX + len(text))) && (i < c.width); i++ {
		c.contents[aY][i] = runeCell{
			character:       rune(text[i-aX]),
			textColor:       c.saveStates[0].textColor,
			backgroundColor: c.saveStates[0].backgroundColor,
		}
	}

	if aX > 0 {
		if (c.contents[aY][aX].backgroundColor != c.contents[aY][aX-1].backgroundColor) ||
			(c.contents[aY][aX].textColor != c.contents[aY][aX-1].textColor) {
			//c.lineBreakage[aY] = true;
			c.setLineBreakage(aY)
		}
	}

	if (aX + len(text)) < c.width-1 {
		if (c.contents[aY][aX+len(text)].backgroundColor != c.contents[aY][aX+len(text)+1].backgroundColor) ||
			(c.contents[aY][aX+len(text)].textColor != c.contents[aY][aX+len(text)+1].textColor) {
			//c.lineBreakage[aY] = true;
			c.setLineBreakage(aY)
		}
	}

}

func (c *Canvas) FillRect(x int, y int, width int, height int) {
	x += c.saveStates[0].translationVector.X
	y += c.saveStates[0].translationVector.Y

	aX := x
	aY := y

	if aX < 0 {
		width += aX
		aX = 0
	}
	if aY < 0 {
		height += aY
		aY = 0
	}

	actualWidth := width
	if (aX + width) > c.width {
		actualWidth = c.width - aX
	}

	if actualWidth < 1 {
		return
	}

	for i := aY; (i < c.height) && i < (height+aY); i++ {

		for a := aX; a < (aX + actualWidth); a++ {
			c.contents[i][a] = runeCell{
				character:       c.saveStates[0].fillChar,
				textColor:       c.saveStates[0].textColor,
				backgroundColor: c.saveStates[0].backgroundColor,
			}
		}

		if aX > 0 {
			if (c.contents[i][aX].backgroundColor != c.contents[i][aX-1].backgroundColor) ||
				(c.contents[i][aX].textColor != c.contents[i][aX-1].textColor) {
				//c.lineBreakage[i] = true;
				c.setLineBreakage(i)
			}
		}

		if (aX + actualWidth - 1) < c.width-1 {
			if (aX+actualWidth < len(c.contents[i])) && (i < len(c.contents)) && (aX+actualWidth > 0) && (i > 0) {
				if (c.contents[i][aX+actualWidth-1].backgroundColor != c.contents[i][aX+actualWidth].backgroundColor) ||
					(c.contents[i][aX+actualWidth-1].textColor != c.contents[i][aX+actualWidth].textColor) {
					//c.lineBreakage[i] = true;
					c.setLineBreakage(i)
				}
			}
		}
	}
}

func (c Canvas) Move(y, x int) {

}

func (c *Canvas) DrawCanvas(x int, y int, can Canvas) {
	x += c.saveStates[0].translationVector.X
	y += c.saveStates[0].translationVector.Y

	aX := x
	aY := y

	if aX < 0 {
		aX = 0
	}
	if aY < 0 {
		aY = 0
	}

	for i := aY; (i < c.height) && ((i - y) < can.height); i++ {
		for a := aX; (a < c.width) && ((a - x) < can.width); a++ {
			//if can.lineBreakage[i-y] {
			if can.hasLineBreakage(i - y) {
				//c.lineBreakage[i] = true;
				c.setLineBreakage(i)
			}
			//c.lineBreakage[i] = can.lineBreakage[i-y];
			c.contents[i][a] = can.contents[i-y][a-x]
		}

		if aX > 0 {
			if (c.contents[i][aX].backgroundColor != c.contents[i][aX-1].backgroundColor) ||
				(c.contents[i][aX].textColor != c.contents[i][aX-1].textColor) {
				//c.lineBreakage[i] = true;
				c.setLineBreakage(i)
			}
		}

		if (aX + can.width) < c.width-1 {
			if (c.contents[i][aX+can.width].backgroundColor != c.contents[i][aX+can.width+1].backgroundColor) ||
				(c.contents[i][aX+can.width].textColor != c.contents[i][aX+can.width+1].textColor) {
				//c.lineBreakage[i] = true;
				c.setLineBreakage(i)
			}
		}
	}
}
