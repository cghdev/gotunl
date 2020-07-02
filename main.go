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

var version = "1.2.2"
var color = map[string]string{
	"red":   "\x1b[31;1m",
	"green": "\x1b[32;1m",
	"reset": "\x1b[0m"}

type connections struct {
	id         int
	name       string
	status     string
	timestamp  int64
	clientAddr string
	serverAddr string
}

func listConnections(gt *gotunl.Gotunl, output string) { // add output format as json?
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
		ptmp.id = p.ID
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
	}
	sort.SliceStable(c, func(i, j int) bool {
		return c[i].id < c[j].id
	})
	table := tablewriter.NewWriter(os.Stdout)
	if anycon {
		table.SetHeader([]string{"ID", "Name", "Status", "Connected for", "Client IP", "Server IP"})
	} else {
		table.SetHeader([]string{"ID", "Name", "Status"})
	}
	setOutputFormat(table, output)
	for _, p := range c {
		since := ""
		if p.timestamp > 0 {
			ts := time.Unix(p.timestamp, 0)
			since = formatSince(ts)
		}
		pid := strconv.Itoa(p.id)
		if anycon {
			table.Append([]string{pid, p.name, p.status, since, p.clientAddr, p.serverAddr})
		} else {
			table.Append([]string{pid, p.name, p.status})
		}
	}
	table.Render()
}

func setOutputFormat(tbl *tablewriter.Table, output string) *tablewriter.Table {
	switch output {
	case "tsv":
		tbl.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		tbl.SetAlignment(tablewriter.ALIGN_LEFT)
		tbl.SetCenterSeparator("")
		tbl.SetColumnSeparator("\t")
		tbl.SetRowSeparator("")
		tbl.SetHeaderLine(false)
		tbl.SetBorder(false)
		tbl.SetTablePadding("\t")
		tbl.SetAutoFormatHeaders(true)
		tbl.SetAutoWrapText(false)
		tbl.SetAutoFormatHeaders(false)
	case "table":
		tbl.SetAutoFormatHeaders(false)
	default:
		fmt.Println("Define output format")
		os.Exit(1)
	}
	return tbl
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
	} else if a.Name == "o" {
		fmt.Printf("  -%v <output>\t%v\n", a.Name, a.Usage)
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
	o := flag.String("o", "table", "Output format table|tsv (default is table)")
	v := flag.Bool("v", false, "Show version")

	flag.Parse()
	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}
	if *l {
		listConnections(&gt, *o)
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
