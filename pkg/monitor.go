package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

type Monitors []struct {
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
	Reserved   []int   `json:"reserved"`
	Scale      float64 `json:"scale"`
	Transform  int     `json:"transform"`
	Focused    bool    `json:"focused"`
	DpmsStatus bool    `json:"dpmsStatus"`
	Vrr        bool    `json:"vrr"`
}

func (m *Monitors) Active() int {
	for _, mon := range *m {
		if mon.Focused {
			return mon.ID
		}
	}
	return 0
}

var MonitorInterfaceName = "com.hypr.monitorService"

type MonitorService struct {
	monitors Monitors
}

// func (MonitorService) parse(msg string) string {
//	parts := strings.Split(msg, ",")
//	if len(parts) != 2 {
//		return ""
//	}
//	log.Print(parts)

//	mon := parts[0]
//	// desktop := parts[1]
//	return fmt.Sprintf(`{ "name": "%s" }`, mon)
// }

func DbusMontitor(ctx context.Context, focusedMonitor chan string) {
	service := MonitorService{}
	conn := GetDbusConnection()
	err := conn.Export(service, ServiceObjectPath, MonitorInterfaceName)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-focusedMonitor:
			log.Print("focusedmon signal received")
			service.monitors = GetMonitors()
			data := fmt.Sprintf(`{ "id": "%v" }`, service.monitors.Active())
			err = conn.Emit(ServiceObjectPath, MonitorInterfaceName+".monitor", data)
			if err != nil {
				log.Print(err)
			}
		}
	}
}

type WorkspaceService struct {
	workspaces Workspaces
}

var WorkspacesInterfaceName = "com.hypr.workspaceService"

func DbusWorkspaces(ctx context.Context, workspace chan string) {
	interfaceName := WorkspacesInterfaceName
	memberName := ".workspaces"

	service := WorkspaceService{
		workspaces: NewWorkspaces(),
	}
	conn := GetDbusConnection()
	err := conn.Export(service, ServiceObjectPath, interfaceName)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-workspace:
			mon := GetMonitors()
			service.workspaces.Update(mon)
			data := service.workspaces.String()
			err = conn.Emit(ServiceObjectPath, interfaceName+memberName, data)
			if err != nil {
				log.Print(err)
			}
		}
	}
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
	Number      string `json:"number"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	monitorID   int
	monitorName string
}

var (
	Active    = Lavender
	Empty     = Bg
	Focused   = []string{Red, Peach, Green, Blue}
	Unfocused = []string{Pink, Yellow, Teal, Lavender}
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

func (w Workspaces) Update(m Monitors) {
	for i := range w {
		w[i].Color = Empty
	}
	// add active montiors
	var MonNames = make(map[string]int)
	for _, mon := range m {
		MonNames[mon.Name] = mon.ID
	}
	// add active workspaces
	for _, ws := range GetHyprWorkspaces() {
		if ws.ID > 0 && ws.ID < 8 {
			w[ws.ID].Color = Unfocused[MonNames[ws.Monitor]]
		}
	}
	for _, mon := range m {
		if mon.ActiveWorkspace.ID <= 7 {
			if mon.Focused {
				w[mon.ActiveWorkspace.ID].Color = Active
			} else {
				w[mon.ActiveWorkspace.ID].Color = Focused[mon.ID]
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
