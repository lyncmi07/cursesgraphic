package cursesgraphic

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
	"github.com/hajimehoshi/ebiten/text"
	"golang.org/x/image/font"
)

const (
	TEXT_WIDTH  = 8
	TEXT_HEIGHT = 16
)

var (
	standardFont font.Face
	realCanvas   *Canvas
	scrWidth     int
	scrHeight    int
)

type mainThread func(screenCanvas *Canvas)

func init() {
	f, err := ebitenutil.OpenFile("cour.ttf")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	tt, err := truetype.Parse(b)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	standardFont = truetype.NewFace(tt, &truetype.Options{
		Size:    14,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	ebiten.SetFullscreen(true)
}

func update(screen *ebiten.Image) error {
	if ebiten.IsRunningSlowly() {
		return nil
	}

	drawCanvasToScreen(screen)

	return nil
}

func formFullStringFromLine(line []runeCell) string {
	fullString := string(line[0].character)

	for i := 1; i < len(line); i++ {
		fullString += string(line[i].character)
	}

	return fullString
}

func formStringFromLine(line []runeCell) (string, int) {
	fullString := string(line[0].character)
	originalTextColor := line[0].textColor
	originalBackgroundColor := line[0].backgroundColor
	size := 1
	for i := 1; i < len(line); i++ {
		if originalTextColor == line[i].textColor &&
			originalBackgroundColor == line[i].backgroundColor {
			fullString += string(line[i].character)
			size++
		} else {
			break
		}
	}

	return fullString, size
}

func drawCanvasToScreen(screen *ebiten.Image) {

	//NOTE:Text can only be rendered on same line, at least one Draw call per line
	//rectangles can obviously be merged for multiple lines however
	//FIXME: Vsync problems. Is it fixable?

	hadLineBreak := true
	currentFirstLineIndex := -1
	screenWidth, _ := screen.Size()

	//print block color
	for i := 0; i < realCanvas.height; i++ {
		if realCanvas.hasLineBreakage(i) {
			if currentFirstLineIndex != -1 {
				//fmt.Printf("Printing for line %d\n", currentFirstLineIndex);
				ebitenutil.DrawRect(screen,
					0,
					float64(currentFirstLineIndex*TEXT_HEIGHT+4),
					float64(screenWidth),
					float64((i-1)*TEXT_HEIGHT+4),
					realCanvas.contents[currentFirstLineIndex][0].backgroundColor)
				hadLineBreak = true
			}
		} else {
			if hadLineBreak {
				currentFirstLineIndex = i
				hadLineBreak = false
			}
		}
	}

	for i := 0; i < realCanvas.height; i++ {
		//if realCanvas.lineBreakage[i] {
		if realCanvas.hasLineBreakage(i) {
			//ensure next line starts a new unbroken line batch
			hadLineBreak = true
			currentLine := realCanvas.contents[i]
			totalLength := len(currentLine)
			for a := 0; a < realCanvas.width; a++ {
				currString, currLength := formStringFromLine(currentLine)
				if currLength == totalLength {
					//realCanvas.lineBreakage[i] = false;
					realCanvas.unsetLineBreakage(i)
				}
				currentLine = currentLine[currLength:]
				ebitenutil.DrawRect(screen,
					float64(a*TEXT_WIDTH),
					(float64(i)*TEXT_HEIGHT + 4),
					float64(scrWidth),
					TEXT_HEIGHT,
					realCanvas.contents[i][a].backgroundColor)
				text.Draw(screen,
					currString,
					standardFont,
					a*TEXT_WIDTH,
					int(float64(i+1)*TEXT_HEIGHT),
					realCanvas.contents[i][a].textColor)

				a += currLength - 1
			}
		} else {
			text.Draw(screen,
				formFullStringFromLine(realCanvas.contents[i]),
				standardFont,
				0,
				int((i+1)*TEXT_HEIGHT),
				realCanvas.contents[currentFirstLineIndex][i].textColor)
		}
	}

	//Display lines on top of characters
	for i := 0; i < len(realCanvas.graphicLines); i++ {
		line := realCanvas.graphicLines[i]
		ebitenutil.DrawLine(screen,
			float64(line.x1*TEXT_WIDTH)+4,
			float64(line.y1)*TEXT_HEIGHT+12,
			float64(line.x2*TEXT_WIDTH)+4,
			float64(line.y2)*TEXT_HEIGHT+12,
			color.RGBA{0xff, 0xff, 0xff, 0xff})
	}

	//Display FPS Counter
	text.Draw(screen,
		fmt.Sprintf("FPS:%.2f", ebiten.CurrentFPS()),
		standardFont,
		0,
		20,
		color.RGBA{0, 0, 0xFF, 0xFF})
}

func CurseGraphicStart(fn mainThread, screenWidth, screenHeight int) {
	scrWidth = screenWidth
	scrHeight = screenHeight

	rc := NewFullscreenCanvas()

	go fn(rc)

	if err := ebiten.Run(update, screenWidth, screenHeight, 1, "Font (Ebiten Demo)"); err != nil {
		log.Fatal(err)
	}
}

func newGenericCanvas() *Canvas {
	rc := new(Canvas)
	rc.saveStates = make([]canContext, 1)
	rc.saveStates[0].fillChar = '.'
	rc.saveStates[0].translationVector = vector{X: 0, Y: 0}

	rc.saveStates[0].backgroundColor = color.RGBA{0, 0, 0, 0xff}
	rc.saveStates[0].textColor = color.RGBA{0xff, 0xff, 0xff, 0xff}

	return rc
}

func newContents(width, height int) [][]runeCell {
	contents := make([][]runeCell, height)
	for i := 0; i < height; i++ {
		contents[i] = make([]runeCell, width)
		for a := 0; a < width; a++ {
			//contents[i][a] = ' ';
			contents[i][a] = runeCell{
				character:       ' ',
				textColor:       color.RGBA{0xff, 0xff, 0xff, 0xff},
				backgroundColor: color.RGBA{0, 0, 0, 0xff},
			}
		}
	}

	return contents
}

func NewFullscreenCanvas() *Canvas {
	rc := newGenericCanvas()

	rc.height = int(float64(scrHeight) / TEXT_HEIGHT)
	rc.width = int(float64(scrWidth) / TEXT_WIDTH)

	rc.lineBreakage = make([]int, 0)
	rc.contents = newContents(rc.width, rc.height)
	realCanvas = rc

	return rc
}

func NewVirtualCanvas(width int, height int) *Canvas {
	vc := newGenericCanvas()
	vc.width = width
	vc.height = height
	vc.lineBreakage = make([]int, 0)
	vc.contents = newContents(vc.width, vc.height)
	return vc
}
