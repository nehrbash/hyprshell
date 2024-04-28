package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/nehrbash/hyprshell/pkg/icon"
)

type Monitor struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Make            string  `json:"make"`
	Model           string  `json:"model"`
	Serial          string  `json:"serial"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	RefreshRate     float64 `json:"refreshRate"`
	X               int     `json:"x"`
	Y               int     `json:"y"`
	ActiveWorkspace struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"activeWorkspace"`
	SpecialWorkspace struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"specialWorkspace"`
	Reserved   []int   `json:"reserved"`
	Scale      float64 `json:"scale"`
	Transform  int     `json:"transform"`
	Focused    bool    `json:"focused"`
	DpmsStatus bool    `json:"dpmsStatus"`
	Vrr        bool    `json:"vrr"`
}

type Monitors []Monitor

func (m *Monitors) Active() Monitor {
	for _, mon := range *m {
		if mon.Focused {
			return mon
		}
	}
	return Monitor{}
}

func GetMonitors() (m Monitors) {
	var outb, errb bytes.Buffer
	cmd := exec.Command("/usr/bin/hyprctl", "monitors", "-j")
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Println("running hyprctl monitors: ", err)
	}
	if err := json.Unmarshal(outb.Bytes(), &m); err != nil {
		log.Println("unmarshal monitors", err)
	}
	return
}

type Workspaces []struct {
	Number           string `json:"number"`
	Color            string `json:"color"`
	Icon             string `json:"icon"`
	ActiveWindowIcon string `json:"active_client_icon"`
	monitorID        int
	monitorName      string
}

var (
	Active          = Lavender
	Empty           = Bg
	FocusedColors   = []string{Red, Peach, Green, Blue}
	UnfocusedColors = []string{Pink, Yellow, Teal, Lavender}
)

func (w Workspaces) String() string {
	// remove ws 0
	w2 := make(Workspaces, len(w))
	copy(w2, w)
	w2 = w2[1:]

	data, err := json.Marshal(w2)
	if err != nil {
		log.Print(err)
	}
	return string(data)
}

func (w Workspaces) Update(m Monitors, clients Clients) {
	for i := range w {
		w[i].Color = Empty
	}
	// add active montiors
	MonNames := make(map[string]int)
	for _, mon := range m {
		MonNames[mon.Name] = mon.ID
	}

	clientIconMap := make(map[string]string)
	for _, c := range clients {
		_, clientIconMap[c.Address] = icon.GetIcon(c.Class)
	}
	// add active workspaces
	for _, ws := range GetHyprWorkspaces() {
		if ws.ID > 0 && ws.ID < 8 {
			w[ws.ID].Color = UnfocusedColors[MonNames[ws.Monitor]]
			w[ws.ID].ActiveWindowIcon = clientIconMap[ws.Lastwindow]
		}
	}
	for _, mon := range m {
		if mon.ActiveWorkspace.ID <= 7 {
			if mon.Focused {
				w[mon.ActiveWorkspace.ID].Color = Active
			} else {
				w[mon.ActiveWorkspace.ID].Color = FocusedColors[mon.ID]
			}
		}
	}
}

func NewWorkspaces() Workspaces {
	w := make(Workspaces, 8)
	for i := 1; i <= 7; i++ {
		w[i].Number = fmt.Sprint(i)
		w[i].Icon = fmt.Sprintf("images/sloth%d.png", i)
		w[i].Color = Empty
	}
	return w
}

type HyprWorkspaces []struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Monitor         string `json:"monitor"`
	Windows         int    `json:"windows"`
	Hasfullscreen   bool   `json:"hasfullscreen"`
	Lastwindow      string `json:"lastwindow"`
	Lastwindowtitle string `json:"lastwindowtitle"`
}

func GetHyprWorkspaces() (w HyprWorkspaces) {
	var outb, errb bytes.Buffer
	cmd := exec.Command("/usr/bin/hyprctl", "workspaces", "-j")
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Println("running hyprctl workspaces", err)
	}
	if err := json.Unmarshal(outb.Bytes(), &w); err != nil {
		log.Println("unmarshal workspaces", err)
	}
	return
}
