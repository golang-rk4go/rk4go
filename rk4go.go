package github.com/golang-rk4go/rk4go
import (
    "fmt"
    "os"
    "os/signal"
    "time"
    "syscall"
    "runtime"
    "sync"
)

type remember struct {
     file string
     stamp int64
     address []int
     amount int
}

const MAX_FILES = 4096
var rkmutex = &sync.Mutex{}
var rklocked bool = false

func RK_CNT(cond bool, t int, f int, m []int) (res bool) {
        res = cond
	if cond {
	   m[t]++;
        } else {
	  m[f]++;
	}
	return cond;
}

func RK_TRUE(cond bool,m []int,t int, i int) (res bool) {
     res = cond;
     if cond {
     	m[t]+=i;
     }
     return;
}

func RK_MI(m []int, t int) (res bool) {
    res = false;
    m[t] = 0;
    return res
}

var book[MAX_FILES] remember;
var current int = 0

func RK_check_in(file string, stamp int64, amount int) (address []int){
       address = make([]int, amount)
       var entry remember
       entry.file = file
       entry.stamp = stamp
       entry.address = address
       entry.amount = amount
       if current==0 {
	        go func() {
		   sigc := make(chan os.Signal, 1)
		   signal.Notify(sigc)
		   for {
		        s := <-sigc
			//fmt.Println("Got signal:", s)
		        RK_check_out();
			if (s==syscall.SIGINT || s==syscall.SIGTERM) {
			   os.Exit(0)
                        }
	           }
                 }()

		go func() {
		   time.Sleep(5 * time.Second)
		   for {
		       RK_check_out();
       		       time.Sleep(20 * time.Second)
                   }
                }();
       }
       if current < MAX_FILES {
	       book[current] = entry  
	       current++
	   } else {
	   fmt.Println("More than ", MAX_FILES, " instrumented, adjust MAX_FILES in rk4go.go");
       } 
       return address
}


func RK_check_out() {
    content:=""
    for k:=0; k < current; k++ {
    	zero := 1
	entry := book[k];
       for ii := 0 ; ii < entry.amount; ii++ {
	   if entry.address[ii] != 0 {
	       zero = 0
	       ii = entry.amount
	   }
        }
	
        if zero == 0 {
	       content+=fmt.Sprintf("{\"time\":%d,",time.Now().Unix())
               hostname, e := os.Hostname()
               if e != nil {
                   hostname = "localhost"
               }
               content+=fmt.Sprintf("\"hostname\":\"%s\",",hostname)
               content+=fmt.Sprintf("\"raw\":\"%s\",",entry.file)
               content+="\"stamp\":0,"
               content+="\"cnt\":["
           first := 0
           for i := 0; i < entry.amount; i++ {
               if first != 0 {
	           content+=","
               } else {
                   first = 1
               }
	       content+=fmt.Sprintf("%d",entry.address[i])
               entry.address[i] = 0
           }
	   content+="]}\n"
	}
    }
    f,e := os.OpenFile("rk-coverage.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    defer f.Close()   
    if e != nil {
    } else {
    
    }
    if runtime.GOOS == "windows" {
    } else { 
        syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
    }
    f.WriteString(content);

}


// atexit → use "C" cgo, callback then to __rk_check_out()
// signal handlers, similar as in C 
