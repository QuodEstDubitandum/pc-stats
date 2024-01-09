package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type stats struct {
	memory float64
	processor float64
	network float64
}

func main() {

	if err := ui.Init(); err != nil {
		fmt.Printf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	ticker := time.NewTicker(time.Second).C

	memory, _ := mem.VirtualMemory()
	processor, _ := cpu.Percent(1000*time.Millisecond, false)

	

	connections, err := net.IOCounters(true)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

	possibleConns := checkForNetworks(connections)
	fmt.Println(len(possibleConns))

	p1 := widgets.NewParagraph()
	p1.SetRect(0, 0, 100, 10)
	p1.Text = fmt.Sprintf("Processor: %v", processor[0])


	p2 := widgets.NewParagraph()
	p2.SetRect(0, 10, 100, 20)
	p2.Text = fmt.Sprintf("Memory: %v", memory.UsedPercent)

	ui.Render(p1, p2)

	for{
		select {
		case e := <-ui.PollEvents():
			if e.ID == "<C-c>" {
				return
			}
			if e.ID == "<Escape>" {
				return
			}
		
		case <-ticker: 
			return
		}
	}
}

func renderFunction(){
	return
}

func checkForNetworks(conns []net.IOCountersStat) []net.IOCountersStat{

	possibleConns := []net.IOCountersStat{}

	for _, conn := range conns {
		if strings.HasPrefix(conn.Name, "wl") || strings.HasPrefix(conn.Name, "en") || strings.HasPrefix(conn.Name, "eth") {
			possibleConns = append(possibleConns, conn)
        // fmt.Printf("Interface Name: %v\n", counter.Name)
        // fmt.Printf("Bytes Sent: %v, Bytes Received: %v\n", counter.BytesSent, counter.BytesRecv)
        // fmt.Printf("Packets Sent: %v, Packets Received: %v\n", counter.PacketsSent, counter.PacketsRecv)
        // fmt.Printf("Error In: %v, Error Out: %v\n", counter.Errin, counter.Errout)
        // fmt.Printf("Drop In: %v, Drop Out: %v\n\n", counter.Dropin, counter.Dropout)
    	}
	}
	fmt.Sprintf("Found %d networks.", len(possibleConns))

	return possibleConns
}