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

type stats struct {
	memory float64
	processor float64
	network float64
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

func spawnRenderRoutine(ctx context.Context){
	ticker := time.NewTicker(time.Second).C

	cpuBuffer := types.NewSlidingWindow[float64](50)
	memBuffer := types.NewSlidingWindow[float64](50)
	networkBuffer := types.NewSlidingWindow[uint64](50)

	for {
		select {
		case <-ticker:
			renderFunction(cpuBuffer, memBuffer, networkBuffer)
		case <-ctx.Done():
			return
		}
	}
}

func renderFunction(cpuBuffer *types.SlidingWindow[float64], memBuffer *types.SlidingWindow[float64], networkBuffer *types.SlidingWindow[uint64]){
	memory, _ := mem.VirtualMemory()
	processor, _ := cpu.Percent(1000*time.Millisecond, false)
	networkConns, _ := net.IOCounters(true)

	cpuBuffer.Push(processor[0])
	memBuffer.Push(memory.UsedPercent)

	p1 := widgets.NewSparkline()
	p1.Data = cpuBuffer.Data
	p1.LineColor = ui.ColorRed

	g1 := widgets.NewSparklineGroup(p1)
	g1.SetRect(0,0,70,10)
	g1.Title = fmt.Sprintf("CPU: %.2f", processor[0]) + "%"
	
	p2 := widgets.NewSparkline()
	p2.Data = memBuffer.Data
	p2.LineColor = ui.ColorRed

	g2 := widgets.NewSparklineGroup(p2)
	g2.SetRect(0,10,70,20)
	g2.Title = fmt.Sprintf("RAM: %.2f", memory.UsedPercent) + "%"

	x := 20

	for _, conn := range networkConns {
		if strings.HasPrefix(conn.Name, "wl") || strings.HasPrefix(conn.Name, "en") || strings.HasPrefix(conn.Name, "eth") {
			networkBuffer.Push(conn.BytesRecv)

			networkTextParagraph := NewBorderlessParagraph()
			networkTextParagraph.SetRect(0, x, 70, x+2)
			networkTextParagraph.Text = fmt.Sprintf("%s: %d", conn.Name, conn.BytesRecv)

			networkGraphParagraph := widgets.NewParagraph()
			networkGraphParagraph.SetRect(0, x+2, 70, x+10)

			ui.Render(networkTextParagraph, networkGraphParagraph)
			x += 10
		}
	}

	t3 := NewBorderlessParagraph()
	t3.SetRect(0, x, 70, x+2)
	t3.Text = "[X] Press Control-C or Escape to quit"

	ui.Render(g1, g2, t3)
	return
}

func renderGraph(){
	return
}