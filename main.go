package main
import (
"fmt"
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
if c.R == 0 { return 0, 0 }
ang := 2.0 * math.Pi * float64(c.I) / float64(8*c.R)
return float64(c.R)*math.Cos(ang), float64(c.R)*math.Sin(ang)
}
var selectedColor = "R"
var currentRadius = 6
var isRunning = false
type clickableRaster struct {
widget.BaseWidget
raster *canvas.Raster
onTap func(fyne.Position)
}
func (r *clickableRaster) CreateRenderer() fyne.WidgetRenderer {
return widget.NewSimpleRenderer(r.raster)
}
func (r *clickableRaster) Tapped(e *fyne.PointEvent) {
if r.onTap != nil { r.onTap(e.Position) }
}
func newClickableRaster(draw func(w, h int) image.Image, tap func(fyne.Position)) *clickableRaster {
r := &clickableRaster{
raster: canvas.NewRaster(draw),
onTap: tap,
}
r.ExtendBaseWidget(r)
return r
}
func parseSource(source string) (map[Cell]string, int) {
grid := make(map[Cell]string)
maxR := 0
for _, line := range strings.Split(source, "\n") {
line = strings.TrimSpace(line)
if strings.HasPrefix(line, "META_RADIUS:") {
p := strings.Split(line, ":")
if len(p) >= 2 { maxR, _ = strconv.Atoi(strings.TrimSpace(p[1])) }
continue
}
if strings.HasPrefix(line, "R") {
p := strings.Split(line, ":")
if len(p) < 2 { continue }
c := strings.Split(p[0][1:], ",")
if len(c) < 2 { continue }
r, _ := strconv.Atoi(c[0])
i, _ := strconv.Atoi(c[1])
grid[Cell{r, i}] = strings.TrimSpace(p[1])
}
}
return grid, maxR
}
func main() {
a := app.New()
w := a.NewWindow("ODL Studio v3.1 - Visual IDE")
w.Resize(fyne.NewSize(1400, 900))
editor := widget.NewMultiLineEntry()
editor.SetText("META_RADIUS: 6\nR1,0: RU\nR3,6: M\nR5,0: H")
logArea := widget.NewMultiLineEntry()
logArea.Disable()
logArea.SetText("=== ODL Console ===\nReady.")
toolbar := container.NewHBox(
widget.NewButton("New", func() {
editor.SetText("META_RADIUS: 6\n")
currentRadius = 6
logArea.SetText("=== ODL Console ===\nNew project created.")
}),
widget.NewButton("Open", func() {
logArea.SetText(logArea.Text + "\n[Open] Coming soon...")
}),
widget.NewButton("Save", func() {
logArea.SetText(logArea.Text + "\n[Save] Coming soon...")
}),
widget.NewSeparator(),
widget.NewButton("▶ Run", func() {
isRunning = true
logArea.SetText(logArea.Text + "\n[Run] Engine not wired yet.")
}),
widget.NewButton("Step", func() {
logArea.SetText(logArea.Text + "\n[Step] Engine not wired yet.")
}),
widget.NewButton("Pause", func() {
isRunning = false
logArea.SetText(logArea.Text + "\n[Pause] Execution paused.")
}),
widget.NewButton("⏹ Reset", func() {
isRunning = false
logArea.SetText("=== ODL Console ===\nReset.")
}),
)
paletteLabel := widget.NewLabel("Color:")
colorsKeys := []string{"R", "B", "Y", "P", "G", "M", "L", "H", "C", "K", "W"}
palette := container.NewHBox()
for _, k := range colorsKeys {
key := k
btn := widget.NewButton(key, func() { selectedColor = key })
palette.Add(btn)
}
radiusSlider := widget.NewSlider(1, 100)
radiusSlider.SetValue(float64(currentRadius))
radiusLabel := widget.NewLabel("Radius: 6")
radiusSlider.OnChanged = func(v float64) {
currentRadius = int(v)
radiusLabel.SetText(fmt.Sprintf("Radius: %d", currentRadius))
lines := strings.Split(editor.Text, "\n")
newLines := []string{fmt.Sprintf("META_RADIUS: %d", currentRadius)}
for _, l := range lines {
if !strings.HasPrefix(l, "META_RADIUS:") { newLines = append(newLines, l) }
}
editor.SetText(strings.Join(newLines, "\n"))
}
var raster *clickableRaster
drawFunc := func(w, h int) image.Image {
img := image.NewRGBA(image.Rect(0, 0, w, h))
for y := 0; y < h; y++ {
for x := 0; x < w; x++ { img.Set(x, y, color.RGBA{20, 20, 25, 255}) }
}
grid, maxR := parseSource(editor.Text)
currentRadius = maxR
halfX, halfY := float64(w)/2.0, float64(h)/2.0
scale := (float64(w) * 0.45) / float64(maxR+1)
if scale > 30 { scale = 30 }
dotSize := int(scale * 0.4)
if dotSize < 1 { dotSize = 1 }
colorMap := map[string]color.RGBA{
"R": {255, 50, 50, 255}, "B": {50, 150, 255, 255}, "Y": {255, 255, 50, 255},
"P": {200, 50, 255, 255}, "G": {50, 255, 50, 255}, "M": {255, 0, 255, 255},
"L": {150, 255, 50, 255}, "H": {150, 0, 255, 255}, "C": {50, 255, 255, 255},
"K": {10, 10, 10, 255}, "W": {200, 200, 200, 255},
}
for r := 0; r <= maxR; r++ {
maxI := 8 * r
if r == 0 { maxI = 1 }
for i := 0; i < maxI; i++ {
c := Cell{r, i}
cx, cy := c.XY()
px, py := int(cx*scale+halfX), int(-cy*scale+halfY)
for dy := -dotSize; dy <= dotSize; dy++ {
for dx := -dotSize; dx <= dotSize; dx++ {
if dx*dx+dy*dy <= dotSize*dotSize {
img.Set(px+dx, py+dy, color.RGBA{50, 50, 60, 255})
}
}
}
s := grid[c]
if s != "" {
col := colorMap[string(s[0])]
for dy := -dotSize; dy <= dotSize; dy++ {
for dx := -dotSize; dx <= dotSize; dx++ {
if dx*dx+dy*dy <= dotSize*dotSize { img.Set(px+dx, py+dy, col) }
}
}
}
}
}
return img
}
tapFunc := func(pos fyne.Position) {
grid, maxR := parseSource(editor.Text)
wSize := raster.Size()
halfX, halfY := float64(wSize.Width)/2.0, float64(wSize.Height)/2.0
scale := (float64(wSize.Width) * 0.45) / float64(maxR+1)
if scale > 30 { scale = 30 }
dx, dy := float64(pos.X)-halfX, halfY-float64(pos.Y)
dist := math.Sqrt(dx*dx + dy*dy)
r := int(math.Round(dist / scale))
if r > maxR { return }
ang := math.Atan2(dy, dx)
if ang < 0 { ang += 2 * math.Pi }
i := int(math.Round(ang * float64(8*r) / (2 * math.Pi)))
if r == 0 { i = 0 } else { i = i % (8 * r) }
target := Cell{r, i}
if selectedColor == "W" { delete(grid, target) } else { grid[target] = selectedColor }
var newLines []string
newLines = append(newLines, fmt.Sprintf("META_RADIUS: %d", maxR))
for c, s := range grid { newLines = append(newLines, fmt.Sprintf("R%d,%d: %s", c.R, c.I, s)) }
editor.SetText(strings.Join(newLines, "\n"))
}
raster = newClickableRaster(drawFunc, tapFunc)
editor.OnChanged = func(s string) { raster.Refresh() }
topControls := container.NewVBox(toolbar, container.NewHBox(radiusLabel, radiusSlider), container.NewHBox(paletteLabel, palette))
leftPanel := container.NewVSplit(editor, logArea)
mainContent := container.NewHSplit(leftPanel, raster)
mainContent.SetOffset(0.35)
w.SetContent(container.NewBorder(topControls, nil, nil, nil, mainContent))
w.ShowAndRun()
}
