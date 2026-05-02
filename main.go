package main
import("fmt";"os";"strings";"strconv";"math")
type Cell struct{R,I int}
func(c Cell)XY()(float64,float64){
if c.R==0{return 0,0}
return float64(c.R)*math.Cos(2.0*math.Pi*float64(c.I)/float64(8*c.R)),float64(c.R)*math.Sin(2.0*math.Pi*float64(c.I)/float64(8*c.R))
}
func Dist(c1,c2 Cell)float64{
x1,y1:=c1.XY()
x2,y2:=c2.XY()
return math.Sqrt((x1-x2)*(x1-x2)+(y1-y2)*(y1-y2))
}
func Neighbors(c Cell,maxR int)[]Cell{
var res []Cell
rMin,rMax:=c.R-1,c.R+1
if rMin<0{rMin=0}
if rMax>maxR{rMax=maxR}
for r:=rMin;r<=rMax;r++{
maxI:=8*r;if r==0{maxI=1}
for i:=0;i<maxI;i++{
nc:=Cell{r,i}
if nc!=c&&Dist(c,nc)<1.5{res=append(res,nc)}
}
}
return res
}
func GetDir(state string)(float64,float64){
if state=="GU"{return 0,1}
if state=="GD"{return 0,-1}
if state=="GL"{return -1,0}
if state=="GR"{return 1,0}
return 0,0
}
func main(){
fmt.Println("ODL v2.0 Phase 2: Neighbors & Vector Movement")
if len(os.Args)<2{return}
data,_:=os.ReadFile(os.Args[1])
grid:=make(map[Cell]string)
maxR:=0
for _,line:=range strings.Split(string(data),"\n"){
line=strings.TrimSpace(line)
if strings.HasPrefix(line,"META_RADIUS:"){maxR,_=strconv.Atoi(strings.Split(line,":")[1]);continue}
if line==""||strings.HasPrefix(line,"#")||strings.HasPrefix(line,"META"){continue}
if strings.HasPrefix(line,"R"){
p:=strings.Split(line,":")
c:=strings.Split(p[0][1:],",")
r,_:=strconv.Atoi(c[0])
i,_:=strconv.Atoi(c[1])
if r>maxR{maxR=r}
grid[Cell{r,i}]=strings.TrimSpace(p[1])
}
}
for c,s:=range grid{
dx,dy:=GetDir(s)
if dx!=0||dy!=0{
var bestCell Cell
var maxDot float64=-999.0
cx,cy:=c.XY()
for _,n:=range Neighbors(c,maxR+1){
nx,ny:=n.XY()
vx,vy:=nx-cx,ny-cy
norm:=math.Sqrt(vx*vx+vy*vy)
if norm>0{vx/=norm;vy/=norm}
dot:=vx*dx+vy*dy
if dot>maxDot{maxDot=dot;bestCell=n}
}
fmt.Printf("Cell R%d,I%d (%s) wants to move to R%d,I%d\n",c.R,c.I,s,bestCell.R,bestCell.I)
}
}
}
