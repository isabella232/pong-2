package main

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/minio/mc/pkg/console"
	"github.com/sparrc/go-ping"
)

func main() {
	down := false
	first := true
	const window = 30
	history := [window]*ping.Statistics{}
	hIndex := window - 1
	rewind := 0
	bolder := color.New(color.Bold)
	console.Printf("%s %s    %s %s   %s %s   %s %s\t%s %s\n", bolder.Sprintf(" \u2261"), bolder.Add(color.FgHiGreen).Sprint("Index"), bolder.Add(color.FgYellow).Sprintf(" \u21de "), bolder.Add(color.FgHiGreen).Sprint("Packets Sent"), bolder.Add(color.FgBlue).Sprintf(" \u21af "), bolder.Add(color.FgHiGreen).Sprint("Packets Recvd."),  bolder.Add(color.FgRed).Sprintf(" \u2691 "), bolder.Add(color.FgHiGreen).Sprint("Packet Loss"), bolder.Sprintf(" \u267e "), bolder.Add(color.FgHiGreen).Sprint("Average Round Trip Time"))
	for {
		pinger, err := ping.NewPinger("8.8.8.8")
		if err != nil {
			console.Fatalln(err)
		}

		finished := make(chan bool, 1)
		pinger.Timeout = 5 * time.Second
		var stats *ping.Statistics
		pinger.OnFinish = func(s *ping.Statistics) {
			stats = s
			finished <- true
		}

		go pinger.Run()

		select {
		case <-time.After(10 * time.Second):
			rewind += 1
			console.Errorln("network latency too high")
			pinger.Stop()
			if !down {
			// 	note := gosxnotifier.NewNotification("High latency detected: Network down")
			// 	note.Title = fmt.Sprintf("latency > 10 seconds")
			// 	note.Sound = gosxnotifier.Default
			// 	if err := note.Push(); err != nil {
			// 		console.Fatalln(err)
			// 	}
			// 	down = true
			}
			continue
		case <-finished:
			if down && !first {
				console.RewindLines(rewind)
				rewind = 0
			}
			down = false
		}
		if !first {
			console.RewindLines(rewind)
			rewind = 0
		}
		hIndex = (hIndex + 1) % window
		history[hIndex] = stats
		first = false
		if stats.PacketLoss > float64(30.00) {
			// note := gosxnotifier.NewNotification("High packet loss detected: Network unstable")
			// note.Title = fmt.Sprintf("Packet loss %0.2f%%", stats.PacketLoss)
			// note.Sound = gosxnotifier.Default
			// if err := note.Push(); err != nil {
			// 	console.Fatalln(err)
			// }
		}
		for i := window; i>0; i-- {
			stat := history[(hIndex + i) % window]
			if stat == nil {
				continue
			}
			rewind ++
			packetLoss := fmt.Sprintf("%0.2f%%", stat.PacketLoss)
			spaces := ""
			for i:=0;i<20-len(packetLoss);i++ {
				spaces = spaces + " "
			}
			formatString := "   %d\t\t%d\t\t   %d  \t\t\t%0.2f%%%s%s \n"
			if stat.PacketLoss > 0.00 {
				clr := color.FgRed
				if stat.PacketLoss < float64(40.00) {
					clr = color.FgYellow
				}
				formatString = color.New(clr).Sprint(formatString)
			}
			console.Printf(formatString, window - i + 1, stat.PacketsSent, stat.PacketsRecv, stat.PacketLoss, spaces, stat.AvgRtt)
		}
		<-time.After(1 * time.Second)
	}
}
