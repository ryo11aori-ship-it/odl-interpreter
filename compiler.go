package main

import (
"fmt"
"os"
"os/exec"
"strings"
)

func main() {
if len(os.Args) < 2 {
fmt.Println("Usage: odlc <source.odl> [output.exe]")
return
}
inFile := os.Args[1]
outFile := strings.TrimSuffix(inFile, ".odl") + ".exe"
if len(os.Args) >= 3 {
outFile = os.Args[2]
}
data, err := os.ReadFile(inFile)
if err != nil {
fmt.Println("Error reading file:", err)
return
}
sourceCode := `package main
import (
"fmt"
"math"
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
data := ` + "`" + string(data) + "`" + `
grid := make(map[Cell]string)
infra := make(map[Cell]string)
maxR := 0
for _, line := range strings.Split(data, "\n") {
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
s := strings.TrimSpace(p[1])
grid[Cell{r, i}] = s
if s == "P" || s == "C" || s == "K" || s == "R" || s == "B" || s == "Y" || s == "O" || s == "Br" || s == "M" || s == "L" || s == "H" {
infra[Cell{r, i}] = s
}
}
}
outBuf := ""
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
if targetInfra == "P" || targetInfra == "K" || targetInfra == "C" || targetInfra == "O" || targetInfra == "Br" || targetInfra == "Y" {
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
}
}
}
for c, s := range infra {
nextBuf[c] = append(nextBuf[c], s)
}
for c, states := range nextBuf {
for i, s := range states {
if len(s) == 2 {
states[i] = Field(c, infra, s, maxR+1)
}
}
}
halt := false
for c, states := range nextBuf {
infraType := infra[c]
if infraType == "M" {
for _, s := range states {
if len(s) == 2 {
if s[0] == 'R' {
outBuf += "1"
} else if s[0] == 'B' {
outBuf += "0"
} else if s[0] == 'G' {
if outBuf != "" {
val, _ := strconv.ParseInt(outBuf, 2, 64)
fmt.Printf("SYSTEM OUTPUT: %c\n", val)
outBuf = ""
}
}
}
}
nextBuf[c] = []string{"M"}
} else if infraType == "H" {
for _, s := range states {
if len(s) == 2 {
halt = true
}
}
nextBuf[c] = []string{"H"}
}
}
grid = make(map[Cell]string)
for c, states := range nextBuf {
grid[c] = Resolve(states)
}
if step%4 == 0 {
for c, s := range infra {
if s == "C" {
var bestN Cell
maxY := -999.0
found := false
for _, n := range Neighbors(c, maxR+1) {
if infra[n] == "" || infra[n] == "R" || infra[n] == "B" || infra[n] == "Y" {
_, ny := n.XY()
if ny > maxY {
maxY = ny
bestN = n
found = true
}
}
}
if found {
if infra[bestN] == "R" {
grid[bestN] = "RU"
} else if infra[bestN] == "B" {
grid[bestN] = "BU"
} else {
grid[bestN] = "GU"
}
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
`
tempFile := "_odl_temp_build.go"
err = os.WriteFile(tempFile, []byte(sourceCode), 0644)
if err != nil {
fmt.Println("Temp File Error:", err)
return
}
defer os.Remove(tempFile)
cmd := exec.Command("go", "build", "-o", outFile, tempFile)
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
err = cmd.Run()
if err != nil {
fmt.Println("Compile Error:", err)
return
}
fmt.Println("Successfully compiled to ->", outFile)
}
