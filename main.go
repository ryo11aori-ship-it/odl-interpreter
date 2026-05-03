package main

import (
"math"
"fyne.io/fyne/v2/app"
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

func Dist(c1, c2 Cell) float64 {
x1, y1 := c1.XY()
x2, y2 := c2.XY()
return math.Sqrt((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2))
}

func Neighbors(c Cell, maxR int) []Cell {
var res []Cell
for r := c.R - 1; r <= c.R+1; r++ {
if r < 0 || r > maxR {
continue
}
maxI := 8 * r
if r == 0 {
maxI = 1
}
for i := 0; i < maxI; i++ {
nc := Cell{r, i}
if nc != c && Dist(c, nc) < 1.5 {
res = append(res, nc)
}
}
}
return res
}

func GetDir(state string) (float64, float64) {
if len(state) == 2 {
if state[1] == 'U' {
return 0, 1
}
if state[1] == 'D' {
return 0, -1
}
if state[1] == 'L' {
return -1, 0
}
if state[1] == 'R' {
return 1, 0
}
}
return 0, 0
}

func Field(c Cell, infra map[Cell]string, s string, maxR int) string {
cx, cy := c.XY()
dx, dy := GetDir(s)
px, py := 0.0, 0.0
found := false
for r := c.R - 2; r <= c.R+2; r++ {
if r < 0 || r > maxR {
continue
}
maxI := 8 * r
if r == 0 {
maxI = 1
}
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
if !found {
return s
}
base := string(s[0])
dirs := []string{base + "U", base + "D", base + "L", base + "R"}
dxs := []float64{0, 0, -1, 1}
dys := []float64{1, -1, 0, 0}
md := -999.0
best := s
for j := 0; j < 4; j++ {
dot := px*dxs[j] + py*dys[j]
if dot > md {
md = dot
best = dirs[j]
}
}
return best
}

func Resolve(states []string) string {
if len(states) == 0 {
return ""
}
var dir string
hasR, hasB, hasY, hasP := false, false, false, false
hasCarrier := false
for _, s := range states {
if s == "R" {
hasR = true
}
if s == "B" {
hasB = true
}
if s == "Y" {
hasY = true
}
if s == "P" {
hasP = true
}
if len(s) == 2 {
hasCarrier = true
dir = string(s[1])
if s[0] == 'R' {
hasR = true
}
if s[0] == 'B' {
hasB = true
}
}
}
if hasP {
return "P"
}
if hasR && hasB && !hasY {
return "Y"
}
if hasR && hasB && hasY {
if hasCarrier {
return "R" + dir
}
return "R"
}
if hasR && hasY {
if hasCarrier {
return "B" + dir
}
return "B"
}
if hasB && hasY {
if hasCarrier {
return "R" + dir
}
return "R"
}
if hasY && hasCarrier {
return "G" + dir
}
if hasY {
return "Y"
}
for _, s := range states {
if s == "O" {
return "O"
}
}
for _, s := range states {
if s == "M" {
return "M"
}
if s == "L" {
return "L"
}
if s == "H" {
return "H"
}
}
if hasCarrier {
if hasR {
return "R" + dir
}
if hasB {
return "B" + dir
}
return "G" + dir
}
if hasR {
return "R"
}
if hasB {
return "B"
}
for _, s := range states {
if s == "C" {
return "C"
}
}
for _, s := range states {
if s == "K" {
return "K"
}
}
for _, s := range states {
if s == "Br" {
return "Br"
}
}
return states[0]
}

func main() {
myApp := app.New()
myWindow := myApp.NewWindow("ODL Studio - Initialization Test")
helloText := widget.NewLabel("Success!\nODL Studio EXE Build is working correctly on your PC.\nPlease tell the AI to proceed to the IDE code.")
myWindow.SetContent(helloText)
myWindow.ShowAndRun()
}
