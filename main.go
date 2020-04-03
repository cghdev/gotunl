package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cghdev/gotunl/pkg/gotunl"
	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
)

var version = "1.2.0"
var color = map[string]string{
	"red":   "\x1b[31;1m",
	"green": "\x1b[32;1m",
	"reset": "\x1b[0m"}

type connections struct {
	id         string
	name       string
	status     string
	timestamp  int64
	clientAddr string
	serverAddr string
}

func listConnections(gt *gotunl.Gotunl) { // add output format as json?
	if len(gt.Profiles) == 0 {
		fmt.Println("No profiles found in Pritunl")
		os.Exit(1)
	}
	cons := gt.GetConnections()
	c := []connections{}
	stdis := ""
	stcon := ""
	anycon := false
	for pid, p := range gt.Profiles {
		ptmp := connections{}
		if runtime.GOOS != "windows" {
			stdis = color["red"] + "Disconnected" + color["reset"]
			stcon = color["green"] + "Connected" + color["reset"]
		} else {
			stdis = "Disconnected"
			stcon = "Connected"
		}
		ptmp.status = stdis
		ptmp.name = gjson.Get(p.Conf, "name").String()
		ptmp.id = strconv.Itoa(p.ID)
		if strings.Contains(cons, pid) {
			ptmp.status = strings.Title(gjson.Get(cons, pid+".status").String())
			ptmp.serverAddr = gjson.Get(cons, pid+".server_addr").String()
			ptmp.clientAddr = gjson.Get(cons, pid+".client_addr").String()
			ptmp.timestamp = gjson.Get(cons, pid+".timestamp").Int()
			if ptmp.status == "Connected" {
				ptmp.status = stcon
				anycon = true
			}
		}
		c = append(c, ptmp)
		sort.Slice(c, func(i, j int) bool {
			return c[i].id < c[j].id
		})
	}
	table := tablewriter.NewWriter(os.Stdout)
	if anycon {
		table.SetHeader([]string{"ID", "Name", "Status", "Connected for", "Client IP", "Server IP"})
	} else {
		table.SetHeader([]string{"ID", "Name", "Status"})
	}
	table.SetAutoFormatHeaders(false)
	for _, p := range c {
		since := ""
		if p.timestamp > 0 {
			ts := time.Unix(p.timestamp, 0)
			since = formatSince(ts)
		}
		if anycon {
			table.Append([]string{p.id, p.name, p.status, since, p.clientAddr, p.serverAddr})
		} else {
			table.Append([]string{p.id, p.name, p.status})
		}
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

func formatSince(t time.Time) string {
	Day := 24 * time.Hour
	ts := time.Since(t)
	sign := time.Duration(1)
	var days, hours, minutes, seconds string
	if ts < 0 {
		sign = -1
		ts = -ts
	}
	d := sign * (ts / Day)
	ts = ts % Day
	h := ts / time.Hour
	ts = ts % time.Hour
	m := ts / time.Minute
	ts = ts % time.Minute
	s := ts / time.Second
	if d > 0 {
		days = fmt.Sprintf("%d days ", d)
	}
	if h > 0 {
		hours = fmt.Sprintf("%d hrs ", h)
	}
	if m > 0 {
		minutes = fmt.Sprintf("%d mins ", m)
	}
	seconds = fmt.Sprintf("%d secs", s)
	return fmt.Sprintf("%v%v%v%v", days, hours, minutes, seconds)
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
