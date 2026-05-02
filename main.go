package main
import("fmt";"os";"strings";"strconv";"math")
type Cell struct{R,I int}
func(c Cell)XY()(float64,float64){
if c.R==0{return 0,0}
return float64(c.R)*math.Cos(2.0*math.Pi*float64(c.I)/float64(8*c.R)),float64(c.R)*math.Sin(2.0*math.Pi*float64(c.I)/float64(8*c.R))
}
func Dist(c1,c2 Cell)float64{
x1,y1:=c1.XY();x2,y2:=c2.XY()
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
func Resolve(states []string)string{
if len(states)==0{return ""}
if len(states)==1{return states[0]}
c:=make(map[string]int)
for _,s:=range states{c[s]++}
if c["R"]>0&&c["B"]>0&&c["Y"]>0{return "R"}
if c["R"]>0&&c["B"]>0&&c["Y"]==0{return "Y"}
if c["R"]>0&&c["Y"]>0{return "B"}
if c["B"]>0&&c["Y"]>0{return "R"}
if c["R"]>0{return "R"}
if c["B"]>0{return "B"}
prio:=map[string]int{"P":90,"Y":80,"O":70,"GU":60,"GD":60,"GL":60,"GR":60,"R":50,"B":50,"C":40,"K":30,"Br":20,"W":10}
b:=""
mp:=0
for _,s:=range states{if prio[s]>mp{mp=prio[s];b=s}}
return b
}
func main(){
fmt.Println("ODL v2.0 Phase 5.1: Collision Fix")
if len(os.Args)<2{return}
data,_:=os.ReadFile(os.Args[1])
grid:=make(map[Cell][]string)
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
grid[Cell{r,i}]=append(grid[Cell{r,i}],strings.TrimSpace(p[1]))
}
}
nextBuf:=make(map[Cell][]string)
for c,states:=range grid{
for _,s:=range states{
if s!="GU"&&s!="GD"&&s!="GL"&&s!="GR"{nextBuf[c]=append(nextBuf[c],s)}
}
}
for c,states:=range grid{
for _,s:=range states{
dx,dy:=GetDir(s)
if dx!=0||dy!=0{
var best Cell
var maxDot float64=-999.0
cx,cy:=c.XY()
for _,n:=range Neighbors(c,maxR+1){
nx,ny:=n.XY()
vx,vy:=nx-cx,ny-cy
norm:=math.Sqrt(vx*vx+vy*vy)
if norm>0{vx/=norm;vy/=norm}
dot:=vx*dx+vy*dy
if dot>maxDot{maxDot=dot;best=n}
}
if maxDot>-999.0{nextBuf[best]=append(nextBuf[best],s)}else{nextBuf[c]=append(nextBuf[c],s)}
}
}
}
finalGrid:=make(map[Cell]string)
for c,states:=range nextBuf{
resolved:=Resolve(states)
if len(states)>1{fmt.Printf("Collision at R%d,I%d : %v -> Resolved to %s\n",c.R,c.I,states,resolved)}
finalGrid[c]=resolved
}
fmt.Println("--- Step 1 Grid State ---")
for k,v:=range finalGrid{fmt.Printf("Cell(R%d, I%d) : %s\n",k.R,k.I,v)}
}
