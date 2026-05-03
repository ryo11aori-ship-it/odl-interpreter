package main
import (
"image"
"image/color"
"math"
"strconv"
"strings"
"fyne.io/fyne/v2"
"fyne.io/fyne/v2/app"
"fyne.io/fyne/v2/canvas"
"fyne.io/fyne/v2/container"
"fyne.io/fyne/v2/widget"
)
type Cell struct {
R int
I int
}
func (c Cell) XY() (float64, float64) {
if c.R == 0 {
return 0, 0
}
ang := 2.0 * math.Pi * float64(c.I) / float64(8*c.R)
return float64(c.R)*math.Cos(ang), float64(c.R)*math.Sin(ang)
}
func drawGrid(w, h int, source string) image.Image {
img := image.NewRGBA(image.Rect(0, 0, w, h))
for y := 0; y < h; y++ {
for x := 0; x < w; x++ {
img.Set(x, y, color.RGBA{25, 25, 30, 255})
}
}
grid := make(map[Cell]string)
maxR := 0
for _, line := range strings.Split(source, "\n") {
line = strings.TrimSpace(line)
if strings.HasPrefix(line, "META_RADIUS:") {
parts := strings.Split(line, ":")
if len(parts) >= 2 {
maxR, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
}
continue
}
if strings.HasPrefix(line, "R") {
p := strings.Split(line, ":")
if len(p) < 2 { continue }
c := strings.Split(p[0][1:], ",")
if len(c) < 2 { continue }
r, _ := strconv.Atoi(c[0])
i, _ := strconv.Atoi(c[1])
if r > maxR { maxR = r }
grid[Cell{r, i}] = strings.TrimSpace(p[1])
}
}
halfX := float64(w) / 2.0
halfY := float64(h) / 2.0
colors := map[string]color.RGBA{
"R": {255, 50, 50, 255}, "B": {50, 150, 255, 255}, "Y": {255, 255, 50, 255},
"P": {200, 50, 255, 255}, "O": {255, 150, 50, 255}, "Br": {139, 69, 19, 255},
"C": {50, 255, 255, 255}, "K": {20, 20, 20, 255}, "W": {255, 255, 255, 255},
"M": {255, 0, 255, 255}, "L": {150, 255, 50, 255}, "H": {150, 0, 255, 255},
}
for r := 0; r <= maxR; r++ {
maxI := 8 * r
if r == 0 { maxI = 1 }
for i := 0; i < maxI; i++ {
c := Cell{r, i}
cx, cy := c.XY()
px := int(cx*30.0 + halfX)
py := int(-cy*30.0 + halfY)
for dy := -8; dy <= 8; dy++ {
for dx := -8; dx <= 8; dx++ {
dist := dx*dx + dy*dy
if dist <= 64 {
if dist > 49 {
img.Set(px+dx, py+dy, color.RGBA{60, 60, 60, 255})
} else {
img.Set(px+dx, py+dy, color.RGBA{40, 40, 40, 255})
}
}
}
}
s := grid[c]
if s != "" {
col, ok := colors[s]
if !ok && len(s) == 2 {
if s[0] == 'G' { col = color.RGBA{50, 255, 50, 255} }
if s[0] == 'R' { col = color.RGBA{255, 50, 50, 255} }
if s[0] == 'B' { col = color.RGBA{50, 150, 255, 255} }
}
for dy := -5; dy <= 5; dy++ {
for dx := -5; dx <= 5; dx++ {
if dx*dx+dy*dy <= 25 {
img.Set(px+dx, py+dy, col)
}
}
}
if len(s) == 2 && (s[0] == 'G' || s[0] == 'R' || s[0] == 'B') {
ddx, ddy := 0, 0
if s[1] == 'U' { ddy = -3 }
if s[1] == 'D' { ddy = 3 }
if s[1] == 'L' { ddx = -3 }
if s[1] == 'R' { ddx = 3 }
for dy := -1; dy <= 1; dy++ {
for dx := -1; dx <= 1; dx++ {
img.Set(px+ddx+dx, py+ddy+dy, color.RGBA{255, 255, 255, 255})
}
}
}
}
}
}
return img
}
func main() {
a := app.New()
w := a.NewWindow("ODL Studio")
w.Resize(fyne.NewSize(1000, 700))
editor := widget.NewMultiLineEntry()
editor.SetText("META_RADIUS: 6\nR1,2: RU\nR0,0: GU\nR1,0: GR\nR3,6: M\nR5,0: H")
raster := canvas.NewRaster(func(w, h int) image.Image {
return drawGrid(w, h, editor.Text)
})
editor.OnChanged = func(s string) {
raster.Refresh()
}
logArea := widget.NewMultiLineEntry()
logArea.Disable()
logArea.SetText("ODL Studio Initialized.\nReady.")
btnRun := widget.NewButton("Run / Pause", func() {
logArea.SetText(logArea.Text + "\n[Run clicked - Engine not yet wired]")
})
btnStep := widget.NewButton("Step", func() {
logArea.SetText(logArea.Text + "\n[Step clicked - Engine not yet wired]")
})
btnCompile := widget.NewButton("Export EXE", func() {
logArea.SetText(logArea.Text + "\n[Export clicked - Packer not yet wired]")
})
controls := container.NewHBox(btnRun, btnStep, btnCompile)
leftPanel := container.NewBorder(controls, nil, nil, nil, container.NewVSplit(editor, logArea))
split := container.NewHSplit(leftPanel, raster)
split.SetOffset(0.3)
w.SetContent(split)
w.ShowAndRun()
}
