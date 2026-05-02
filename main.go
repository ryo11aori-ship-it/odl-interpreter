package main

import (
"fmt"
"image"
"image/color"
"image/png"
"math"
"os"
"strconv"
"strings"
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

func Field(c Cell, buf map[Cell][]string, s string, maxR int) string {
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
hasP := false
for _, st := range buf[nc] {
if st == "P" {
hasP = true
}
}
if hasP {
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
c := make(map[string]int)
var dir string
hasR, hasB, hasY, hasP := false, false, false, false
for _, s := range states {
c[s]++
if s == "R" || (len(s) == 2 && s[0] == 'R') {
hasR = true
}
if s == "B" || (len(s) == 2 && s[0] == 'B') {
hasB = true
}
if s == "Y" {
hasY = true
}
if s == "P" {
hasP = true
}
if len(s) == 2 && (s[0] == 'G' || s[0] == 'R' || s[0] == 'B') {
dir = string(s[1])
}
}
if hasP {
return "P"
}
if hasR && hasB && hasY {
if dir != "" {
return "R" + dir
}
return "R"
}
if hasR && hasB && !hasY {
return "Y"
}
if hasR && hasY {
if dir != "" {
return "B" + dir
}
return "B"
}
if hasB && hasY {
if dir != "" {
return "R" + dir
}
return "R"
}
if hasY {
return "Y"
}
for _, s := range states {
if s == "O" {
return "O"
}
}
if dir != "" {
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

func Draw(grid map[Cell]string, maxR, step int) {
size := (maxR + 2) * 60
half := float64(size) / 2.0
img := image.NewRGBA(image.Rect(0, 0, size, size))
for y := 0; y < size; y++ {
for x := 0; x < size; x++ {
img.Set(x, y, color.RGBA{25, 25, 30, 255})
}
}
colors := map[string]color.RGBA{
"R": {255, 50, 50, 255},
"B": {50, 150, 255, 255},
"Y": {255, 255, 50, 255},
"P": {200, 50, 255, 255},
"O": {255, 150, 50, 255},
"Br": {139, 69, 19, 255},
"C": {50, 255, 255, 255},
"K": {20, 20, 20, 255},
"W": {255, 255, 255, 255},
}
for r := 0; r <= maxR; r++ {
maxI := 8 * r
if r == 0 {
maxI = 1
}
for i := 0; i < maxI; i++ {
c := Cell{r, i}
cx, cy := c.XY()
px := int(cx*30.0 + half)
py := int(-cy*30.0 + half)
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
if !ok {
if len(s) == 2 {
if s[0] == 'G' {
col = color.RGBA{50, 255, 50, 255}
}
if s[0] == 'R' {
col = color.RGBA{255, 50, 50, 255}
}
if s[0] == 'B' {
col = color.RGBA{50, 150, 255, 255}
}
}
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
if s[1] == 'U' {
ddy = -3
}
if s[1] == 'D' {
ddy = 3
}
if s[1] == 'L' {
ddx = -3
}
if s[1] == 'R' {
ddx = 3
}
for dy := -1; dy <= 1; dy++ {
for dx := -1; dx <= 1; dx++ {
img.Set(px+ddx+dx, py+ddy+dy, color.RGBA{255, 255, 255, 255})
}
}
}
}
}
}
f, _ := os.Create(fmt.Sprintf("step_%02d.png", step))
png.Encode(f, img)
f.Close()
}

func main() {
fmt.Println("ODL v2.0 Phase 10: Data Transport & Logic Gates")
if len(os.Args) < 2 {
return
}
data, _ := os.ReadFile(os.Args[1])
grid := make(map[Cell]string)
maxR := 0
buf := make(map[Cell][]string)
for _, line := range strings.Split(string(data), "\n") {
line = strings.TrimSpace(line)
if strings.HasPrefix(line, "META_RADIUS:") {
maxR, _ = strconv.Atoi(strings.Split(line, ":")[1])
continue
}
if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "META") {
continue
}
if strings.HasPrefix(line, "R") {
p := strings.Split(line, ":")
c := strings.Split(p[0][1:], ",")
r, _ := strconv.Atoi(c[0])
i, _ := strconv.Atoi(c[1])
if r > maxR {
maxR = r
}
buf[Cell{r, i}] = append(buf[Cell{r, i}], strings.TrimSpace(p[1]))
}
}
for c, states := range buf {
grid[c] = Resolve(states)
}
Draw(grid, maxR, 0)
for step := 1; step <= 20; step++ {
nextBuf := make(map[Cell][]string)
for c, s := range grid {
dx, dy := GetDir(s)
if dx != 0 || dy != 0 {
var best Cell
maxDot := -999.0
cx, cy := c.XY()
for _, n := range Neighbors(c, maxR+1) {
targetState := grid[n]
if targetState == "P" || targetState == "K" || targetState == "C" {
continue
}
nx, ny := n.XY()
vx, vy := nx-cx, ny-cy
norm := math.Sqrt(vx*vx + vy*vy)
if norm > 0 {
vx /= norm
vy /= norm
}
dot := vx*dx + vy*dy
if dot > maxDot {
maxDot = dot
best = n
}
}
if maxDot > -999.0 {
nextBuf[best] = append(nextBuf[best], s)
} else {
nextBuf[c] = append(nextBuf[c], s)
}
} else {
nextBuf[c] = append(nextBuf[c], s)
}
}
for c, states := range nextBuf {
for i, s := range states {
if len(s) == 2 && (s[0] == 'G' || s[0] == 'R' || s[0] == 'B') {
states[i] = Field(c, nextBuf, s, maxR+1)
}
}
}
grid = make(map[Cell]string)
for c, states := range nextBuf {
grid[c] = Resolve(states)
}
for c, s := range grid {
if s == "Br" {
for _, n := range Neighbors(c, maxR+1) {
ns := grid[n]
if ns == "R" || ns == "B" || strings.HasPrefix(ns, "G") {
grid[n] = ""
}
}
}
}
for c, s := range grid {
if s == "O" {
for _, n := range Neighbors(c, maxR+1) {
ns := grid[n]
if ns == "R" || ns == "B" {
for _, nn := range Neighbors(c, maxR+1) {
if grid[nn] == "" {
grid[nn] = ns
break
}
}
break
}
}
}
}
if step%4 == 0 {
for c, s := range grid {
if s == "C" {
var bestN Cell
maxY := -999.0
found := false
for _, n := range Neighbors(c, maxR+1) {
if grid[n] == "" {
_, ny := n.XY()
if ny > maxY {
maxY = ny
bestN = n
found = true
}
}
}
if found {
grid[bestN] = "GU"
}
}
}
}
Draw(grid, maxR, step)
fmt.Printf("Step %02d completed.\n", step)
}
}
