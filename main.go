package main
import("fmt";"os";"strings";"strconv")
type Cell struct{R,I int}
func main(){
fmt.Println("ODL v2.0 Engine Initialization...")
if len(os.Args)<2{fmt.Println("Usage: odl <file>");return}
data,err:=os.ReadFile(os.Args[1])
if err!=nil{fmt.Println("Error:",err);return}
grid:=make(map[Cell]string)
lines:=strings.Split(string(data),"\n")
for _,line:=range lines{
line=strings.TrimSpace(line)
if line==""||strings.HasPrefix(line,"#")||strings.HasPrefix(line,"META"){continue}
if strings.HasPrefix(line,"R"){
parts:=strings.Split(line,":")
coords:=strings.Split(parts[0][1:],",")
r,_:=strconv.Atoi(coords[0])
i,_:=strconv.Atoi(coords[1])
grid[Cell{r,i}]=strings.TrimSpace(parts[1])
}
}
fmt.Println("--- Initial Grid State ---")
for k,v:=range grid{fmt.Printf("Cell(R%d, I%d) : %s\n",k.R,k.I,v)}
}
