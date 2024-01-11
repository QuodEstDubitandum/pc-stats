package main

import (
	"context"
	"fmt"
	"pc-stats-cli/types"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type NetworkStats struct {
	Name string
	BytesReceived *types.SlidingWindow[float64]
	BytesSent *types.SlidingWindow[float64]
	LastBytesRec float64
	LastBytesSent float64
	MaxBytesRec float64
	MaxBytesSent float64
}

type BorderlessWidget struct {
	Block  		ui.Block
	Text 		string
	TextStyle 	ui.Style
	WrapText  	bool
}

func main() {
	if err := ui.Init(); err != nil {
		fmt.Printf("failed to initialize termui: %v", err)
	}
	defer ui.Close()


	ctx, cancel := context.WithCancel(context.Background())
	go spawnRenderRoutine(ctx)

	for{
		select {
		case e := <-ui.PollEvents():
			if e.ID == "<C-c>" {
				cancel()
				return
			}
			if e.ID == "<Escape>" {
				cancel()
				return
			}
		}
	}
}

func spawnRenderRoutine(ctx context.Context){
	ticker := time.NewTicker(time.Second).C

	cpuBuffer := types.NewSlidingWindow[float64](50)
	memBuffer := types.NewSlidingWindow[float64](50)

	var networks []*NetworkStats
	allNetworks, _ := net.IOCounters(true)

	for i, conn := range allNetworks {
		if strings.HasPrefix(conn.Name, "wl") || strings.HasPrefix(conn.Name, "en") || strings.HasPrefix(conn.Name, "eth") {
			networks = append(networks, &NetworkStats{
				Name: conn.Name,
				BytesReceived: types.NewSlidingWindow[float64](50),
				BytesSent: types.NewSlidingWindow[float64](50),
				LastBytesRec: float64(conn.BytesRecv),
				LastBytesSent: float64(conn.BytesSent),
				MaxBytesRec: 0,
				MaxBytesSent: 0,
			})
			networks[i-1].BytesReceived.Push(0)
			networks[i-1].BytesSent.Push(0)
		}
	}

	for {
		select {
		case <-ticker:
			renderFunction(cpuBuffer, memBuffer, networks)
		case <-ctx.Done():
			return
		}
	}
}

func renderFunction(cpuBuffer *types.SlidingWindow[float64], memBuffer *types.SlidingWindow[float64], networks []*NetworkStats){
	memory, _ := mem.VirtualMemory()
	processor, _ := cpu.Percent(1000*time.Millisecond, false)

	cpuBuffer.Push(processor[0])
	memBuffer.Push(memory.UsedPercent)

	p1 := widgets.NewSparkline()
	p1.Data = cpuBuffer.Data
	p1.LineColor = ui.ColorGreen

	g1 := widgets.NewSparklineGroup(p1)
	g1.SetRect(0,0,70,10)
	g1.Title = fmt.Sprintf("CPU: %.2f", processor[0]) + "%"
	g1.TitleStyle.Bg = ui.ColorBlack
	g1.TitleStyle.Fg = ui.ColorGreen
	g1.BorderStyle.Fg = ui.ColorRed
	
	p2 := widgets.NewSparkline()
	p2.Data = memBuffer.Data
	p2.LineColor = ui.ColorBlue

	g2 := widgets.NewSparklineGroup(p2)
	g2.SetRect(0,10,70,20)
	g2.Title = fmt.Sprintf("RAM: %.2f", memory.UsedPercent) + "%"
	g2.TitleStyle.Bg = ui.ColorBlack
	g2.TitleStyle.Fg = ui.ColorBlue
	g2.BorderStyle.Fg = ui.ColorRed

	x := 20
	renderNetworkGraph(networks, &x)

	t3 := NewBorderlessParagraph()
	t3.SetRect(0, x, 70, x+2)
	t3.Text = "[X] Press Control-C or Escape to quit"

	ui.Render(g1, g2, t3)
	return
}

func renderNetworkGraph(networks []*NetworkStats, x *int){
	var bytesRecDiff float64
	var bytesSentDiff float64

	allNetworks, _ := net.IOCounters(true)

	for i, network := range networks {
		for _, conn := range allNetworks {
			if conn.Name == network.Name {
				bytesRecDiff = float64(conn.BytesRecv) - network.LastBytesRec
				network.BytesReceived.Push(bytesRecDiff)
				network.LastBytesRec = float64(conn.BytesRecv)
				if bytesRecDiff > network.MaxBytesRec {
					network.MaxBytesRec = bytesRecDiff
				}

				bytesSentDiff = float64(conn.BytesSent) - network.LastBytesSent
				network.BytesSent.Push(bytesSentDiff)
				network.LastBytesSent = float64(conn.BytesSent)
				if bytesSentDiff > network.MaxBytesSent {
					network.MaxBytesSent = bytesSentDiff
				}

				n1 := widgets.NewSparkline()
				n1.Data = network.BytesReceived.Data
				n1.LineColor = ui.ColorCyan
				n1.Title = determineUIText(bytesRecDiff, network.MaxBytesRec, "Download")
				n1.TitleStyle.Bg = ui.ColorBlack
				n1.TitleStyle.Fg = ui.ColorCyan

				n2 := widgets.NewSparkline()
				n2.Data = network.BytesSent.Data
				n2.LineColor = ui.ColorCyan
				n2.Title = determineUIText(bytesSentDiff, network.MaxBytesSent, "Upload")
				n2.TitleStyle.Bg = ui.ColorBlack
				n2.TitleStyle.Fg = ui.ColorCyan

				ng1 := widgets.NewSparklineGroup(n1, n2)
				ng1.SetRect(0, *x, 70, *x+15)
				ng1.Title = fmt.Sprintf("Network %d: %s", i+1, network.Name)
				ng1.TitleStyle.Bg = ui.ColorBlack
				ng1.TitleStyle.Fg = ui.ColorCyan
				ng1.BorderStyle.Fg = ui.ColorRed

				*x += 15
				ui.Render(ng1)
			}
		}
		
	}
	return
}

func determineUIText(bytes float64, maxBytes float64, direction string) string{
	var maxBytesText string
	if maxBytes > 1000000000 {
		maxBytesText = fmt.Sprintf("Max: %.2f GB/s", maxBytes/1000000000)
	}else if maxBytes > 1000000 {
		maxBytesText = fmt.Sprintf("Max: %.2f MB/s", maxBytes/1000000)
	}else{
		maxBytesText = fmt.Sprintf("Max: %.2f KB/s", maxBytes/1000)
	}

	// GB
	if bytes > 1000000000 {
		return fmt.Sprintf("%s: %.2f GB/s || %s", direction, bytes/1000000000, maxBytesText)
	}

	// MB
	if bytes > 1000000 {
		return fmt.Sprintf("%s: %.2f MB/s || %s", direction, bytes/1000000, maxBytesText)
	}

	// KB
	return fmt.Sprintf("%s: %.2f KB/s || %s", direction, bytes/1000, maxBytesText)
}


func NewBorderlessParagraph() *widgets.Paragraph {
	return &widgets.Paragraph{
		Block: *NewBorderlessBlock(),
		TextStyle: ui.Theme.Paragraph.Text,
		WrapText: true,
	}
} 

func NewBorderlessBlock() *ui.Block {
	return &ui.Block{
		Border:       false,
		BorderStyle:  ui.Theme.Block.Border,
		BorderLeft:   false,
		BorderRight:  false,
		BorderTop:    false,
		BorderBottom: false,
		PaddingTop: 2,
		PaddingLeft: 2,
	}
}