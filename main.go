package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/cghdev/gotunl"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
)

var version = "1.0.0"
var color = map[string]string{
	"red":   "\x1b[31;1m",
	"green": "\x1b[32;1m",
	"reset": "\x1b[0m"}

type connections [][]string

func (c connections) Len() int           { return len(c) }
func (c connections) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c connections) Less(i, j int) bool { return c[i][0] < c[j][0] }

func listConnections(gt *gotunl.Gotunl) { // add output format as json?
	if len(gt.Profiles) == 0 {
		fmt.Println("No profiles found in Pritunl")
		os.Exit(1)
	}
	cons := gt.GetConnections()
	c := connections{}
	stdis := ""
	stcon := ""
	for pid, p := range gt.Profiles {
		if runtime.GOOS != "windows" {
			stdis = color["red"] + "Disconnected" + color["reset"]
			stcon = color["green"] + "Connected" + color["reset"]
		} else {
			stdis = "Disconnected"
			stcon = "Connected"
		}
		status := stdis
		if strings.Contains(cons, pid) {
			status = strings.Title(gjson.Get(cons, pid+".status").String())
			if status == "Connected" {
				status = stcon
			}
		}
		ptmp := []string{strconv.Itoa(p.ID), gjson.Get(p.Conf, "name").String(), status}
		c = append(c, ptmp)
		sort.Sort(c)
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Status"})
	table.SetAutoFormatHeaders(false)
	for _, p := range c {
		table.Append(p)
	}
	table.Render()
}

func disconnect(gt *gotunl.Gotunl, id string) {
	if id == "all" {
		gt.StopConnections()
	} else {
		for pid, p := range gt.Profiles {
			if id == gjson.Get(p.Conf, "name").String() || id == strconv.Itoa(p.ID) {
				gt.DisconnectProfile(pid)
			}
		}
	}

}

func connect(gt *gotunl.Gotunl, id string) {
	for pid, p := range gt.Profiles {
		if id == gjson.Get(p.Conf, "name").String() || id == strconv.Itoa(p.ID) {
			gt.ConnectProfile(pid, "", "")
		}
	}
}

func usage(a *flag.Flag) {
	if a.Name == "l" || a.Name == "v" {
		fmt.Printf("  -%v \t\t%v\n", a.Name, a.Usage)
	} else {
		fmt.Printf("  -%v <profile>\t%v\n", a.Name, a.Usage)
	}
}

func main() {
	gt := *gotunl.New()
	flag.Usage = func() {
		fmt.Print("Pritunl command line client\n\n")
		fmt.Println("Usage:")
		flag.VisitAll(usage)
	}
	l := flag.Bool("l", false, "List connections")
	c := flag.String("c", "", "Connect to profile ID or Name")
	d := flag.String("d", "", "Disconnect profile or \"all\"")
	v := flag.Bool("v", false, "Show version")

	flag.Parse()
	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}
	if *l {
		listConnections(&gt)
		os.Exit(0)
	}
	if *c != "" {
		connect(&gt, *c)
		os.Exit(0)
	}
	if *d != "" {
		disconnect(&gt, *d)
		os.Exit(0)
	}
	if *v {
		fmt.Println(version)
	}
	os.Exit(1)
}
