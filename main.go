package main
import (
"bytes"
"encoding/binary"
"fmt"
"image"
"image/color"
"io"
"math"
"net/http"
"os"
"os/exec"
"strconv"
"strings"
"sync"
"time"
"fyne.io/fyne/v2"
"fyne.io/fyne/v2/app"
"fyne.io/fyne/v2/canvas"
"fyne.io/fyne/v2/container"
"fyne.io/fyne/v2/dialog"
"fyne.io/fyne/v2/storage"
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
func Dist(c1, c2 Cell) float64 {
x1, y1 := c1.XY()
x2, y2 := c2.XY()
return math.Sqrt((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2))
}
func Neighbors(c Cell, maxR int) []Cell {
var res []Cell
for r := c.R - 1; r <= c.R+1; r++ {
if r < 0 || r > maxR { continue }
maxI := 8 * r
if r == 0 { maxI = 1 }
for i := 0; i < maxI; i++ {
nc := Cell{r, i}
if nc != c && Dist(c, nc) < 1.5 { res = append(res, nc) }
}
}
return res
}
func GetDir(state string) (float64, float64) {
if len(state) == 2 {
if state[1] == 'U' { return 0, 1 }
if state[1] == 'D' { return 0, -1 }
if state[1] == 'L' { return -1, 0 }
if state[1] == 'R' { return 1, 0 }
}
return 0, 0
}
func Field(c Cell, infra map[Cell]string, s string, maxR int) string {
cx, cy := c.XY()
dx, dy := GetDir(s)
px, py := 0.0, 0.0
found := false
for r := c.R - 2; r <= c.R+2; r++ {
if r < 0 || r > maxR { continue }
maxI := 8 * r
if r == 0 { maxI = 1 }
for i := 0; i < maxI; i++ {
nc := Cell{r, i}
if infra[nc] == "P" {
nx, ny := nc.XY()
vx, vy := nx-cx, ny-cy
if vx*dx+vy*dy >= -0.1 {
d := math.Sqrt(vx*vx + vy*vy)
if d > 0 && d < 2.5 {
px += vx / d
py += vy / d
found = true
}
}
}
}
}
if !found { return s }
base := string(s[0])
dirs := []string{base + "U", base + "D", base + "L", base + "R"}
dxs := []float64{0, 0, -1, 1}
dys := []float64{1, -1, 0, 0}
md := -999.0
best := s
for j := 0; j < 4; j++ {
dot := px*dxs[j] + py*dys[j]
if dot > md { md = dot; best = dirs[j] }
}
return best
}
func Resolve(states []string) string {
if len(states) == 0 { return "" }
var dir string
hasR, hasB, hasY, hasP := false, false, false, false
hasCarrier := false
for _, s := range states {
if s == "R" { hasR = true }
if s == "B" { hasB = true }
if s == "Y" { hasY = true }
if s == "P" { hasP = true }
if len(s) == 2 {
hasCarrier = true
dir = string(s[1])
if s[0] == 'R' { hasR = true }
if s[0] == 'B' { hasB = true }
}
}
if hasP { return "P" }
if hasR && hasB && !hasY { return "Y" }
if hasR && hasB && hasY {
if hasCarrier { return "R" + dir }
return "R"
}
if hasR && hasY {
if hasCarrier { return "B" + dir }
return "B"
}
if hasB && hasY {
if hasCarrier { return "R" + dir }
return "R"
}
if hasY && hasCarrier { return "G" + dir }
if hasY { return "Y" }
for _, s := range states {
if s == "O" { return "O" }
}
for _, s := range states {
if s == "M" { return "M" }
if s == "L" { return "L" }
if s == "H" { return "H" }
if s == "F" { return "F" }
if s == "N" { return "N" }
if s == "X" { return "X" }
}
if hasCarrier {
if hasR { return "R" + dir }
if hasB { return "B" + dir }
return "G" + dir
}
if hasR { return "R" }
if hasB { return "B" }
for _, s := range states {
if s == "C" { return "C" }
}
for _, s := range states {
if s == "K" { return "K" }
}
for _, s := range states {
if s == "Br" { return "Br" }
}
return states[0]
}
func binToStr(bin string) string {
res := ""
for i := 0; i < len(bin); i += 8 {
end := i + 8
if end > len(bin) { end = len(bin) }
val, _ := strconv.ParseInt(bin[i:end], 2, 64)
res += string(rune(val))
}
return res
}
func runODL(data string) {
grid := make(map[Cell]string)
infra := make(map[Cell]string)
maxR := 0
for _, line := range strings.Split(data, "\n") {
line = strings.TrimSpace(line)
if strings.HasPrefix(line, "META_RADIUS:") {
parts := strings.Split(line, ":")
if len(parts) >= 2 { maxR, _ = strconv.Atoi(strings.TrimSpace(parts[1])) }
continue
}
if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "META") { continue }
if strings.HasPrefix(line, "R") {
p := strings.Split(line, ":")
if len(p) < 2 { continue }
c := strings.Split(p[0][1:], ",")
if len(c) < 2 { continue }
r, _ := strconv.Atoi(c[0])
i, _ := strconv.Atoi(c[1])
if r > maxR { maxR = r }
s := strings.TrimSpace(p[1])
grid[Cell{r, i}] = s
if s == "P" || s == "C" || s == "K" || s == "R" || s == "B" || s == "Y" || s == "O" || s == "Br" || s == "M" || s == "L" || s == "H" || s == "F" || s == "N" || s == "X" {
infra[Cell{r, i}] = s
}
}
}
outBuf := ""
bufF := ""
bufN := ""
bufX := ""
for step := 1; step <= 500; step++ {
nextBuf := make(map[Cell][]string)
for c, s := range grid {
if len(s) == 2 {
dx, dy := GetDir(s)
if dx != 0 || dy != 0 {
var best Cell
maxDot := -999.0
cx, cy := c.XY()
for _, n := range Neighbors(c, maxR+1) {
targetInfra := infra[n]
if targetInfra == "P" || targetInfra == "K" || targetInfra == "C" || targetInfra == "O" || targetInfra == "Br" || targetInfra == "Y" { continue }
nx, ny := n.XY()
vx, vy := nx-cx, ny-cy
norm := math.Sqrt(vx*vx + vy*vy)
if norm > 0 { vx /= norm; vy /= norm }
dot := vx*dx + vy*dy
if dot > maxDot { maxDot = dot; best = n }
}
if maxDot > -999.0 {
nextBuf[best] = append(nextBuf[best], s)
} else {
nextBuf[c] = append(nextBuf[c], s)
}
}
}
}
for c, s := range infra {
nextBuf[c] = append(nextBuf[c], s)
}
for c, states := range nextBuf {
for i, s := range states {
if len(s) == 2 { states[i] = Field(c, infra, s, maxR+1) }
}
}
halt := false
for c, states := range nextBuf {
infraType := infra[c]
if infraType == "M" {
for _, s := range states {
if len(s) == 2 {
if s[0] == 'R' { outBuf += "1" } else if s[0] == 'B' { outBuf += "0" } else if s[0] == 'G' {
if outBuf != "" {
val, _ := strconv.ParseInt(outBuf, 2, 64)
fmt.Printf("SYSTEM OUTPUT: %c\n", val)
outBuf = ""
}
}
}
}
nextBuf[c] = []string{"M"}
} else if infraType == "F" {
for _, s := range states {
if len(s) == 2 {
if s[0] == 'R' { bufF += "1" } else if s[0] == 'B' { bufF += "0" } else if s[0] == 'G' {
if bufF != "" {
str := binToStr(bufF)
fmt.Printf("SYSTEM FILE WRITE: %s\n", str)
f, _ := os.OpenFile("odl_output.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
if f != nil { f.WriteString(str + "\n"); f.Close() }
bufF = ""
}
}
}
}
nextBuf[c] = []string{"F"}
} else if infraType == "N" {
for _, s := range states {
if len(s) == 2 {
if s[0] == 'R' { bufN += "1" } else if s[0] == 'B' { bufN += "0" } else if s[0] == 'G' {
if bufN != "" {
str := binToStr(bufN)
fmt.Printf("SYSTEM NETWORK GET: %s\n", str)
go http.Get(str)
bufN = ""
}
}
}
}
nextBuf[c] = []string{"N"}
} else if infraType == "X" {
for _, s := range states {
if len(s) == 2 {
if s[0] == 'R' { bufX += "1" } else if s[0] == 'B' { bufX += "0" } else if s[0] == 'G' {
if bufX != "" {
str := binToStr(bufX)
fmt.Printf("SYSTEM EXECUTE: %s\n", str)
exec.Command("cmd", "/c", str).Start()
bufX = ""
}
}
}
}
nextBuf[c] = []string{"X"}
} else if infraType == "H" {
for _, s := range states {
if len(s) == 2 { halt = true }
}
nextBuf[c] = []string{"H"}
}
}
grid = make(map[Cell]string)
for c, states := range nextBuf { grid[c] = Resolve(states) }
if step%4 == 0 {
for c, s := range infra {
if s == "C" {
var bestN Cell
maxY := -999.0
found := false
for _, n := range Neighbors(c, maxR+1) {
if infra[n] == "" || infra[n] == "R" || infra[n] == "B" || infra[n] == "Y" {
_, ny := n.XY()
if ny > maxY { maxY = ny; bestN = n; found = true }
}
}
if found {
if infra[bestN] == "R" { grid[bestN] = "RU" } else if infra[bestN] == "B" { grid[bestN] = "BU" } else { grid[bestN] = "GU" }
}
}
}
}
if halt {
fmt.Println("PROGRAM HALTED.")
break
}
}
}
var (
selectedColor = "R"
currentRadius = 6
isRunning = false
isUserTyping = true
panX float64 = 0
panY float64 = 0
zoomLevel float64 = 1.0
mu sync.Mutex
engineGrid map[Cell]string
engineInfra map[Cell]string
engineMaxR int
stepCount int
outBufUI string
bufFUI string
bufNUI string
bufXUI string
)
type clickableRaster struct {
widget.BaseWidget
raster *canvas.Raster
onTap func(fyne.Position)
onDrag func(*fyne.DragEvent)
onScroll func(*fyne.ScrollEvent)
}
func (r *clickableRaster) CreateRenderer() fyne.WidgetRenderer { return widget.NewSimpleRenderer(r.raster) }
func (r *clickableRaster) Tapped(e *fyne.PointEvent) {
if r.onTap != nil { r.onTap(e.Position) }
}
func (r *clickableRaster) Dragged(e *fyne.DragEvent) {
if r.onDrag != nil { r.onDrag(e) }
}
func (r *clickableRaster) DragEnd() {}
func (r *clickableRaster) Scrolled(e *fyne.ScrollEvent) {
if r.onScroll != nil { r.onScroll(e) }
}
func newClickableRaster(draw func(w, h int) image.Image, tap func(fyne.Position)) *clickableRaster {
r := &clickableRaster{ raster: canvas.NewRaster(draw), onTap: tap }
r.ExtendBaseWidget(r)
return r
}
func parseSource(source string) {
mu.Lock()
defer mu.Unlock()
engineGrid = make(map[Cell]string)
engineInfra = make(map[Cell]string)
engineMaxR = 6
for _, line := range strings.Split(source, "\n") {
line = strings.TrimSpace(line)
if strings.HasPrefix(line, "META_RADIUS:") {
p := strings.Split(line, ":")
if len(p) >= 2 { engineMaxR, _ = strconv.Atoi(strings.TrimSpace(p[1])) }
continue
}
if strings.HasPrefix(line, "R") {
p := strings.Split(line, ":")
if len(p) < 2 { continue }
c := strings.Split(p[0][1:], ",")
if len(c) < 2 { continue }
r, _ := strconv.Atoi(c[0])
i, _ := strconv.Atoi(c[1])
if r > engineMaxR { engineMaxR = r }
s := strings.TrimSpace(p[1])
engineGrid[Cell{r, i}] = s
if s == "P" || s == "C" || s == "K" || s == "R" || s == "B" || s == "Y" || s == "O" || s == "Br" || s == "M" || s == "L" || s == "H" || s == "F" || s == "N" || s == "X" {
engineInfra[Cell{r, i}] = s
}
}
}
currentRadius = engineMaxR
stepCount = 0
outBufUI = ""
bufFUI = ""
bufNUI = ""
bufXUI = ""
}
func main() {
magicStr := strings.Join([]string{"!!!ODL", "PACKER", "BOUNDARY!!!"}, "_")
magic := []byte(magicStr)
exePath, err := os.Executable()
if err == nil {
exeData, err := os.ReadFile(exePath)
if err == nil {
idx := bytes.LastIndex(exeData, magic)
if idx != -1 {
sourceData := string(exeData[idx+len(magic):])
runODL(sourceData)
return
}
}
}
a := app.New()
w := a.NewWindow("ODL Studio v3.1 - Visual IDE")
w.Resize(fyne.NewSize(1400, 900))
editor := widget.NewMultiLineEntry()
editor.SetText("META_RADIUS: 6\nR1,0: RU\nR3,6: M\nR5,0: H")
logArea := widget.NewMultiLineEntry()
logArea.SetText("=== ODL Console ===\nReady.")
parseSource(editor.Text)
engineStep := func() {
mu.Lock()
defer mu.Unlock()
stepCount++
nextBuf := make(map[Cell][]string)
for c, s := range engineGrid {
if len(s) == 2 {
dx, dy := GetDir(s)
if dx != 0 || dy != 0 {
var best Cell
maxDot := -999.0
cx, cy := c.XY()
for _, n := range Neighbors(c, engineMaxR+1) {
targetInfra := engineInfra[n]
if targetInfra == "P" || targetInfra == "K" || targetInfra == "C" || targetInfra == "O" || targetInfra == "Br" || targetInfra == "Y" { continue }
nx, ny := n.XY()
vx, vy := nx-cx, ny-cy
norm := math.Sqrt(vx*vx + vy*vy)
if norm > 0 { vx /= norm; vy /= norm }
dot := vx*dx + vy*dy
if dot > maxDot { maxDot = dot; best = n }
}
if maxDot > -999.0 {
nextBuf[best] = append(nextBuf[best], s)
} else {
nextBuf[c] = append(nextBuf[c], s)
}
}
}
}
for c, s := range engineInfra { nextBuf[c] = append(nextBuf[c], s) }
for c, states := range nextBuf {
for i, s := range states {
if len(s) == 2 { states[i] = Field(c, engineInfra, s, engineMaxR+1) }
}
}
halt := false
for c, states := range nextBuf {
infraType := engineInfra[c]
if infraType == "M" {
for _, s := range states {
if len(s) == 2 {
if s[0] == 'R' { outBufUI += "1" } else if s[0] == 'B' { outBufUI += "0" } else if s[0] == 'G' {
if outBufUI != "" {
val, _ := strconv.ParseInt(outBufUI, 2, 64)
logArea.SetText(logArea.Text + fmt.Sprintf("\nSYSTEM OUTPUT: %c", val))
outBufUI = ""
}
}
}
}
nextBuf[c] = []string{"M"}
} else if infraType == "F" {
for _, s := range states {
if len(s) == 2 {
if s[0] == 'R' { bufFUI += "1" } else if s[0] == 'B' { bufFUI += "0" } else if s[0] == 'G' {
if bufFUI != "" {
str := binToStr(bufFUI)
logArea.SetText(logArea.Text + fmt.Sprintf("\n[SYSTEM FILE] Wrote: %s", str))
f, _ := os.OpenFile("odl_output.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
if f != nil { f.WriteString(str + "\n"); f.Close() }
bufFUI = ""
}
}
}
}
nextBuf[c] = []string{"F"}
} else if infraType == "N" {
for _, s := range states {
if len(s) == 2 {
if s[0] == 'R' { bufNUI += "1" } else if s[0] == 'B' { bufNUI += "0" } else if s[0] == 'G' {
if bufNUI != "" {
str := binToStr(bufNUI)
logArea.SetText(logArea.Text + fmt.Sprintf("\n[SYSTEM NET] GET: %s", str))
go http.Get(str)
bufNUI = ""
}
}
}
}
nextBuf[c] = []string{"N"}
} else if infraType == "X" {
for _, s := range states {
if len(s) == 2 {
if s[0] == 'R' { bufXUI += "1" } else if s[0] == 'B' { bufXUI += "0" } else if s[0] == 'G' {
if bufXUI != "" {
str := binToStr(bufXUI)
logArea.SetText(logArea.Text + fmt.Sprintf("\n[SYSTEM EXEC] %s", str))
exec.Command("cmd", "/c", str).Start()
bufXUI = ""
}
}
}
}
nextBuf[c] = []string{"X"}
} else if infraType == "H" {
for _, s := range states {
if len(s) == 2 { halt = true }
}
nextBuf[c] = []string{"H"}
}
}
engineGrid = make(map[Cell]string)
for c, states := range nextBuf { engineGrid[c] = Resolve(states) }
if stepCount%4 == 0 {
for c, s := range engineInfra {
if s == "C" {
var bestN Cell
maxY := -999.0
found := false
for _, n := range Neighbors(c, engineMaxR+1) {
if engineInfra[n] == "" || engineInfra[n] == "R" || engineInfra[n] == "B" || engineInfra[n] == "Y" {
_, ny := n.XY()
if ny > maxY { maxY = ny; bestN = n; found = true }
}
}
if found {
if engineInfra[bestN] == "R" { engineGrid[bestN] = "RU" } else if engineInfra[bestN] == "B" { engineGrid[bestN] = "BU" } else { engineGrid[bestN] = "GU" }
}
}
}
}
if halt {
isRunning = false
logArea.SetText(logArea.Text + "\n[!] PROGRAM HALTED.")
}
}
var raster *clickableRaster
toolbar := container.NewHBox(
widget.NewButton("New", func() {
isUserTyping = false
editor.SetText("META_RADIUS: 6\n")
isUserTyping = true
parseSource(editor.Text)
panX = 0
panY = 0
zoomLevel = 1.0
if raster != nil { raster.Refresh() }
logArea.SetText("=== ODL Console ===\nNew project created.")
}),
widget.NewButton("Open", func() {
fd := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
if err != nil || r == nil { return }
defer r.Close()
data, _ := io.ReadAll(r)
isUserTyping = false
editor.SetText(string(data))
isUserTyping = true
parseSource(string(data))
panX = 0
panY = 0
zoomLevel = 1.0
if raster != nil { raster.Refresh() }
logArea.SetText(logArea.Text + "\n[Open] Loaded file: " + r.URI().Name())
}, w)
fd.SetFilter(storage.NewExtensionFileFilter([]string{".odl", ".txt"}))
fd.Show()
}),
widget.NewButton("Save", func() {
fd := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
if err != nil || wc == nil { return }
defer wc.Close()
wc.Write([]byte(editor.Text))
logArea.SetText(logArea.Text + "\n[Save] Saved file: " + wc.URI().Name())
}, w)
fd.SetFileName("circuit.odl")
fd.Show()
}),
widget.NewButton("Export EXE", func() {
exePath, err := os.Executable()
if err != nil {
logArea.SetText(logArea.Text + "\n[Export] Error locating executable.")
return
}
exeData, err := os.ReadFile(exePath)
if err != nil {
logArea.SetText(logArea.Text + "\n[Export] Error reading executable.")
return
}
if len(exeData) > 0x40 {
e_lfanew := binary.LittleEndian.Uint32(exeData[0x3C:0x40])
subsystemOffset := e_lfanew + 92
if subsystemOffset < uint32(len(exeData)) {
exeData[subsystemOffset] = 3
}
}
magicStr := strings.Join([]string{"!!!ODL", "PACKER", "BOUNDARY!!!"}, "_")
magic := []byte(magicStr)
outData := append(exeData, magic...)
outData = append(outData, []byte(editor.Text)...)
err = os.WriteFile("exported_program.exe", outData, 0755)
if err != nil {
logArea.SetText(logArea.Text + "\n[Export] Error writing file.")
return
}
logArea.SetText(logArea.Text + "\n[Export] Successfully packed to -> exported_program.exe")
}),
widget.NewSeparator(),
widget.NewButton("▶ Run", func() {
if !isRunning {
parseSource(editor.Text)
isRunning = true
logArea.SetText(logArea.Text + "\n[Run] Engine started.")
}
}),
widget.NewButton("Step", func() {
isRunning = false
engineStep()
logArea.SetText(logArea.Text + "\n[Step] Advanced 1 tick.")
}),
widget.NewButton("Pause", func() {
isRunning = false
logArea.SetText(logArea.Text + "\n[Pause] Execution paused.")
}),
widget.NewButton("⏹ Reset", func() {
isRunning = false
parseSource(editor.Text)
panX = 0
panY = 0
zoomLevel = 1.0
if raster != nil { raster.Refresh() }
logArea.SetText("=== ODL Console ===\nReset to initial state.")
}),
)
paletteLabel := widget.NewLabel("Color:")
colorsKeys := []string{"R", "B", "Y", "P", "G", "M", "L", "H", "C", "K", "W", "F", "N", "X"}
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
}
drawFunc := func(w, h int) image.Image {
mu.Lock()
defer mu.Unlock()
img := image.NewRGBA(image.Rect(0, 0, w, h))
for y := 0; y < h; y++ {
for x := 0; x < w; x++ { img.Set(x, y, color.RGBA{20, 20, 25, 255}) }
}
halfX, halfY := float64(w)/2.0+panX, float64(h)/2.0+panY
baseScale := (float64(w) * 0.45) / float64(engineMaxR+1)
if baseScale > 30 { baseScale = 30 }
scale := baseScale * zoomLevel
dotSize := int(scale * 0.4)
if dotSize < 1 { dotSize = 1 }
colorMap := map[string]color.RGBA{
"R": {255, 50, 50, 255}, "B": {50, 150, 255, 255}, "Y": {255, 255, 50, 255},
"P": {200, 50, 255, 255}, "G": {50, 255, 50, 255}, "M": {255, 0, 255, 255},
"L": {150, 255, 50, 255}, "H": {150, 0, 255, 255}, "C": {50, 255, 255, 255},
"K": {10, 10, 10, 255}, "W": {200, 200, 200, 255},
"F": {0, 150, 200, 255}, "N": {255, 120, 0, 255}, "X": {220, 20, 60, 255},
}
for r := 0; r <= engineMaxR; r++ {
maxI := 8 * r
if r == 0 { maxI = 1 }
for i := 0; i < maxI; i++ {
c := Cell{r, i}
cx, cy := c.XY()
px, py := int(cx*scale+halfX), int(-cy*scale+halfY)
if px < -50 || px > w+50 || py < -50 || py > h+50 { continue }
s := engineGrid[c]
if s == "" {
pxVec, pyVec := 0.0, 0.0
foundP := false
for r2 := c.R - 2; r2 <= c.R+2; r2++ {
if r2 < 0 || r2 > engineMaxR { continue }
maxI2 := 8 * r2; if r2 == 0 { maxI2 = 1 }
for i2 := 0; i2 < maxI2; i2++ {
if engineInfra[Cell{r2, i2}] == "P" {
nx, ny := Cell{r2, i2}.XY()
vx, vy := nx-cx, ny-cy
d := math.Sqrt(vx*vx + vy*vy)
if d > 0 && d < 2.5 {
pxVec += vx / d; pyVec += vy / d; foundP = true
}
}
}
}
for dy := -dotSize; dy <= dotSize; dy++ {
for dx := -dotSize; dx <= dotSize; dx++ {
if dx*dx+dy*dy <= dotSize*dotSize { img.Set(px+dx, py+dy, color.RGBA{50, 50, 60, 255}) }
}
}
if foundP {
vLen := math.Sqrt(pxVec*pxVec + pyVec*pyVec)
if vLen > 0 {
dxVis := int((pxVec / vLen) * float64(dotSize) * 1.5)
dyVis := int(-(pyVec / vLen) * float64(dotSize) * 1.5)
for dy := -1; dy <= 1; dy++ {
for dx := -1; dx <= 1; dx++ {
img.Set(px+dxVis+dx, py+dyVis+dy, color.RGBA{200, 50, 255, 180})
}
}
}
}
} else {
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
mu.Lock()
wSize := raster.Size()
halfX, halfY := float64(wSize.Width)/2.0+panX, float64(wSize.Height)/2.0+panY
baseScale := (float64(wSize.Width) * 0.45) / float64(engineMaxR+1)
if baseScale > 30 { baseScale = 30 }
scale := baseScale * zoomLevel
dx, dy := float64(pos.X)-halfX, halfY-float64(pos.Y)
dist := math.Sqrt(dx*dx + dy*dy)
r := int(math.Round(dist / scale))
if r > engineMaxR { mu.Unlock(); return }
ang := math.Atan2(dy, dx)
if ang < 0 { ang += 2 * math.Pi }
i := int(math.Round(ang * float64(8*r) / (2 * math.Pi)))
if r == 0 { i = 0 } else { i = i % (8 * r) }
target := Cell{r, i}
if selectedColor == "W" {
delete(engineGrid, target)
delete(engineInfra, target)
} else {
engineGrid[target] = selectedColor
if selectedColor == "P" || selectedColor == "C" || selectedColor == "K" || selectedColor == "R" || selectedColor == "B" || selectedColor == "Y" || selectedColor == "O" || selectedColor == "Br" || selectedColor == "M" || selectedColor == "L" || selectedColor == "H" || selectedColor == "F" || selectedColor == "N" || selectedColor == "X" {
engineInfra[target] = selectedColor
}
}
var newLines []string
newLines = append(newLines, fmt.Sprintf("META_RADIUS: %d", engineMaxR))
for c, s := range engineGrid { newLines = append(newLines, fmt.Sprintf("R%d,%d: %s", c.R, c.I, s)) }
mu.Unlock()
isUserTyping = false
editor.SetText(strings.Join(newLines, "\n"))
isUserTyping = true
}
raster = newClickableRaster(drawFunc, tapFunc)
raster.onDrag = func(e *fyne.DragEvent) {
panX += float64(e.Dragged.DX)
panY += float64(e.Dragged.DY)
raster.Refresh()
}
raster.onScroll = func(e *fyne.ScrollEvent) {
zoomLevel += float64(e.Scrolled.DY) * 0.1
if zoomLevel < 0.1 { zoomLevel = 0.1 }
if zoomLevel > 10.0 { zoomLevel = 10.0 }
raster.Refresh()
}
editor.OnChanged = func(s string) {
if isUserTyping && !isRunning { parseSource(s) }
}
go func() {
for {
if isRunning { engineStep() }
raster.Refresh()
time.Sleep(100 * time.Millisecond)
}
}()
topControls := container.NewVBox(toolbar, container.NewHBox(radiusLabel, radiusSlider), container.NewHBox(paletteLabel, palette))
leftPanel := container.NewVSplit(editor, logArea)
mainContent := container.NewHSplit(leftPanel, raster)
mainContent.SetOffset(0.35)
w.SetContent(container.NewBorder(topControls, nil, nil, nil, mainContent))
w.ShowAndRun()
}
