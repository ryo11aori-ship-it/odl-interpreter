package main
import("fmt";"os";"strings";"strconv";"math")
type Cell struct{R,I int}
func(c Cell)XY()(float64,float64){
if c.R==0{return 0,0}
ang:=2.0*math.Pi*float64(c.I)/float64(8*c.R)
return float64(c.R)*math.Cos(ang),float64(c.R)*math.Sin(ang)
}
func main(){
fmt.Println("ODL v2.0 Phase 1: Coordinate & Step Engine")
if len(os.Args)<2{return}
data,_:=os.ReadFile(os.Args[1])
grid:=make(map[Cell]string)
for _,line:=range strings.Split(string(data),"\n"){
line=strings.TrimSpace(line)
if line==""||strings.HasPrefix(line,"#")||strings.HasPrefix(line,"META"){continue}
if strings.HasPrefix(line,"R"){
p:=strings.Split(line,":")
c:=strings.Split(p[0][1:],",")
r,_:=strconv.Atoi(c[0])
i,_:=strconv.Atoi(c[1])
grid[Cell{r,i}]=strings.TrimSpace(p[1])
}
}
nextGrid:=make(map[Cell]string)
for cell,state:=range grid{
x,y:=cell.XY()
fmt.Printf("Cell R%d,I%d (State:%s) -> X:%.2f, Y:%.2f\n",cell.R,cell.I,state,x,y)
if state=="GU"{fmt.Println("-> Action: GU must move UP in next step")}
nextGrid[cell]=state
}
fmt.Println("--- Next Step Processed ---")
}
