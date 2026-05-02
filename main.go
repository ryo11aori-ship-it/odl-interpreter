package main
import("fmt";"os";"strings")
func main(){
fmt.Println("ODL v2.0 Engine Initialization...")
if len(os.Args)>1{
data,err:=os.ReadFile(os.Args[1])
if err!=nil{fmt.Println("File Error:",err);return}
lines:=strings.Split(string(data),"\n")
for _,line:=range lines{
line=strings.TrimSpace(line)
if line==""||strings.HasPrefix(line,"#"){continue}
fmt.Println("Parsed IR ->",line)
}
}else{
fmt.Println("Usage: odl-repl.exe <source.odl>")
}
}
