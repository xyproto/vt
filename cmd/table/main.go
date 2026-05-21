package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/xyproto/vt"
)

type user struct {
	First, Last, Username, Email, Phone string
}

var users = []user{
	{"Marcus", "Chen", "mchen42", "marcus.chen@techwave.io", "+1-415-555-0147"},
	{"Aisha", "Patel", "aishap", "", "(212)555-0893"},
	{"Oscar", "Lindqvist", "oscarlind", "oscar@lindqvist.se", ""},
	{"Fatima", "Al-Rashid", "fatimar", "fatima.r@openstack.dev", "+44-20-7946-0958"},
	{"Tomás", "Rivera", "trivera", "", ""},
	{"Ingrid", "Bergström", "ibergstrom", "ingrid.b@nordicsoft.fi", "+358-9-555-1234"},
	{"Dmitri", "Volkov", "dvolkov", "", "555-0162"},
	{"Yuki", "Tanaka", "yukitan", "y.tanaka@matsuri.jp", "+81-3-5555-7890"},
	{"Priya", "Sharma", "psharma", "priya@devhub.in", ""},
	{"Leo", "Dubois", "leodub", "", "+33-1-55-42-68-00"},
	{"Amara", "Okafor", "aokafor", "amara.okafor@nexgen.ng", ""},
	{"Sven", "Eriksson", "svene", "sven.eriksson@volvo.se", "+46-31-555-4422"},
	{"Mei", "Zhang", "meizhang", "", "(628)555-0371"},
	{"Rafael", "Costa", "rcosta", "rafael.costa@empresa.br", ""},
	{"Hannah", "Müller", "hmueller", "h.mueller@autobahn.de", "+49-30-555-8899"},
	{"Kofi", "Mensah", "kmensah", "", ""},
	{"Elena", "Popescu", "elenap", "elena.p@bucharest.tech", "+40-21-555-6677"},
	{"Liam", "O'Brien", "lobrien", "liam@greenfield.ie", "+353-1-555-2233"},
	{"Sakura", "Mori", "sakuram", "", "+81-6-5555-4321"},
	{"Carlos", "Vega", "cvega", "carlos.vega@andino.cl", ""},
}

func main() {
	tty, err := vt.NewTTY()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer tty.Close()

	vt.Init()
	defer vt.Close()

	c := vt.NewCanvas()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)

	headers := []string{"First", "Last", "Username", "Email", "Phone"}
	selectedRow := 0
	scrollOffset := 0

	headerBg := vt.TrueBackground(32, 32, 48)
	headerFg := vt.TrueColor(255, 255, 255)
	selBg := vt.TrueBackground(32, 64, 255)
	row1Bg := vt.TrueBackground(8, 8, 8)
	row2Bg := vt.TrueBackground(16, 16, 16)
	fg := vt.TrueColor(200, 200, 200)
	titleFg := vt.TrueColor(64, 128, 255)
	statusFg := vt.TrueColor(180, 180, 180)

	render := func() {
		c.Clear()
		w, h := int(c.W()), int(c.H())

		// Title
		title := fmt.Sprintf("  Table Demo — %d users  [↑/↓ navigate, q quit]", len(users))
		c.Write(0, 0, titleFg.Bold(), headerBg, title)
		// Fill title row
		for col := vt.StringWidth(title); col < w; col++ {
			c.WriteRuneB(uint(col), 0, titleFg, headerBg, ' ')
		}

		// Headers at row 2
		headerRow := uint(2)
		cols := len(headers)
		colW := max(w/cols, 5)
		for col := range w {
			c.WriteRuneB(uint(col), headerRow, headerFg, headerBg, ' ')
		}
		for ci, hdr := range headers {
			x := ci * colW
			for j, r := range hdr {
				if x+j >= w {
					break
				}
				c.WriteRuneB(uint(x+j), headerRow, headerFg.Bold(), headerBg, r)
			}
		}

		// Table rows
		tableStart := 3
		visibleRows := h - tableStart - 1
		if scrollOffset > selectedRow {
			scrollOffset = selectedRow
		}
		if selectedRow >= scrollOffset+visibleRows {
			scrollOffset = selectedRow - visibleRows + 1
		}

		for vi := 0; vi < visibleRows && scrollOffset+vi < len(users); vi++ {
			ri := scrollOffset + vi
			u := users[ri]
			row := uint(tableStart + vi)
			fields := []string{u.First, u.Last, u.Username, u.Email, u.Phone}

			bg := row1Bg
			if ri%2 == 1 {
				bg = row2Bg
			}
			if ri == selectedRow {
				bg = selBg
			}

			// Fill row
			for col := range w {
				c.WriteRuneB(uint(col), row, fg, bg, ' ')
			}

			for ci, field := range fields {
				if field == "" {
					field = "—"
				}
				x := ci * colW
				maxLen := colW - 1
				if len(field) > maxLen {
					field = field[:maxLen-1] + "…"
				}
				for j, r := range field {
					if x+j >= w {
						break
					}
					c.WriteRuneB(uint(x+j), row, fg, bg, r)
				}
			}
		}

		// Status bar
		status := fmt.Sprintf(" Row %d/%d", selectedRow+1, len(users))
		statusRow := uint(h - 1)
		for col := range w {
			c.WriteRuneB(uint(col), statusRow, statusFg, headerBg, ' ')
		}
		c.Write(0, statusRow, statusFg, headerBg, status)

		c.HideCursorAndDraw()
	}

	render()

	for {
		select {
		case <-sigCh:
			if nc := c.Resized(); nc != nil {
				c = nc
			}
			render()
			continue
		default:
		}

		key := tty.ReadKey()
		switch key {
		case "q", "c:3", "c:17":
			return
		case "k", "↑":
			if selectedRow > 0 {
				selectedRow--
			}
		case "j", "↓":
			if selectedRow < len(users)-1 {
				selectedRow++
			}
		case "⇞":
			selectedRow -= 10
			if selectedRow < 0 {
				selectedRow = 0
			}
		case "⇟":
			selectedRow += 10
			if selectedRow >= len(users) {
				selectedRow = len(users) - 1
			}
		case "g":
			selectedRow = 0
		case "G":
			selectedRow = len(users) - 1
		}
		render()
	}
}
