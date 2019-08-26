package main

import (
	"fmt"
	"time"
	"runtime"
	"io"
	"os"
	"os/exec"
	
	"github.com/fatih/color"
	"github.com/minio/mc/pkg/console"
	"github.com/sparrc/go-ping"
	"github.com/spf13/cobra"
)

var pong = &cobra.Command{
	Use: "pong",
	Short: "summarized ping",
	SilenceErrors: true,
	SilenceUsage: true,
	Run: run,
}

var privileged bool
var install bool

func init() {
	pong.Flags().BoolVarP(&privileged, "privileged", "p", false, "run in privileged mode")
	pong.Flags().BoolVarP(&install, "install", "i", false, "install pong in $PATH")
}

func main() {
	pong.Execute()
}

func run(c *cobra.Command, args []string) {
	if install {
		switch runtime.GOOS {
		case "linux":
			exe, err := os.Executable()
			if err != nil {
				console.Fatalln(err)
			}
			r, err := os.Open(exe)
			if err != nil {
				console.Fatalln(err)
			}
			defer r.Close()
			w, err := os.OpenFile("/usr/local/bin/pong", os.O_CREATE|os.O_WRONLY, 0777)
			if err != nil {
				console.Fatalln(err)
			}
			defer w.Close()
			if _, err := io.Copy(w, r); err != nil {
				console.Fatalln(err)
			}
			path, err := exec.LookPath("setcap")
			if err != nil {
				console.Fatalln(err)
			}
			cmd := exec.Command(path,[]string{"cap_net_raw=+ep", "/usr/local/bin/pong"}...)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				os.Remove("/usr/local/bin/pong")	
				console.Fatalln(err)
			}
			os.Remove(exe)
			console.Infof("pong succesfully installed to /usr/local/bin/pong")
		case "darwin":
			exe, err := os.Executable()
			if err != nil {
				console.Fatalln(err)
			}
			r, err := os.Open(exe)
			if err != nil {
				console.Fatalln(err)
			}
			defer r.Close()
			w, err := os.OpenFile("/usr/local/bin/pong", os.O_CREATE|os.O_WRONLY, 0777)
			if err != nil {
				console.Fatalln(err)
			}
			defer w.Close()
			if _, err := io.Copy(w, r); err != nil {
				console.Fatalln(err)
			}
		case "windows":
			console.Fatalln("not yet implemented")
		} 
	}
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
		if privileged || runtime.GOOS == "linux" {
			pinger.SetPrivileged(true)
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
