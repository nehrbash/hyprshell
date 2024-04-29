package core

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/nehrbash/hyprshell/pkg/icon"
)

type (
	Clients []Client
	Client  struct {
		Address   string `json:"address"`
		Mapped    bool   `json:"mapped"`
		Hidden    bool   `json:"hidden"`
		At        []int  `json:"at"`
		Size      []int  `json:"size"`
		Workspace struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"workspace"`
		Floating       bool   `json:"floating"`
		Monitor        int    `json:"monitor"`
		Class          string `json:"class"`
		Title          string `json:"title"`
		Pid            int    `json:"pid"`
		Xwayland       bool   `json:"xwayland"`
		Pinned         bool   `json:"pinned"`
		Fullscreen     bool   `json:"fullscreen"`
		FullscreenMode int    `json:"fullscreenMode"`
		FakeFullscreen bool   `json:"fakeFullscreen"`
		Grouped        []any  `json:"grouped"`
		Swallowing     string `json:"swallowing"`
	}
)

type App struct {
	ID        int
	Title     string
	Class     string
	Color     string
	Address   string
	updated   bool
	Focused   bool
	Swallowed bool
}

type HyprSignal struct {
	Type string
	Msg  string
}

type HyprSignalManager struct {
	HyprEvent       chan HyprSignal
	DockSignal      chan string
	SubmapSignal    chan string
	WeatherSignal   chan string
	MonitorSignal   chan string
	WorkspaceSignal chan string
	QuoteSignal     chan string
	TitleSignal     chan string
}

func (h *HyprSignalManager) HyprListen(ctx context.Context) {
	c, err := net.Dial("unix", os.ExpandEnv(xdg.RuntimeDir+"/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket2.sock"))
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	reader := bufio.NewReader(c)
	lreader := textproto.NewReader(reader)
	for {
		line, err := lreader.ReadLine()
		if err != nil {
			continue
		}
		eventMsg := strings.Split(string(line), ">>")
		if len(eventMsg) != 2 {
			continue
		}

		h.HyprEvent <- HyprSignal{Type: eventMsg[0], Msg: eventMsg[1]}
	}
}

func (h *HyprSignalManager) HyprClientListen(ctx context.Context) {
	for {
		select {
		case hyperEvent := <-h.HyprEvent:
			switch event := hyperEvent.Type; event {
			case "activewindow", "closewindow":
				if event == "closewindow" {
					// TODO seems to need delay before calling hyprctl
					time.Sleep(time.Millisecond * 300)
				}
				h.DockSignal <- "update"
				h.WorkspaceSignal <- hyperEvent.Msg
			case "submap":
				h.SubmapSignal <- hyperEvent.Msg
			case "focusedmon":
				h.MonitorSignal <- hyperEvent.Msg
				h.WorkspaceSignal <- hyperEvent.Msg
			case "destroyworkspace", "createworkspace", "workspace":
				h.WorkspaceSignal <- hyperEvent.Msg
			}
		case <-ctx.Done():
			return
		}
	}
}

func GetClients() (clients Clients) {
	var outb, errb bytes.Buffer
	cmd := exec.Command("/usr/bin/hyprctl", "clients", "-j")
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}
	if err := json.Unmarshal(outb.Bytes(), &clients); err != nil {
		log.Println(err)
	}
	return
}

type AppClass struct {
	ID          int
	Class       string
	Apps        []*App
	Focused     bool
	Favorite    bool
	Desktop     string
	Icon        string
	LastFocused string // last app of this class that was focused
}

// AddCnt is a temp solution to making apps sortable.
// TODO they now have an ID
var (
	AddCnt   int
	classCnt int
)

var activeColor string = "active"

type Apps struct {
	lookup  map[string]*App      // address key
	byClass map[string]*AppClass // className key
}

func NewApps() Apps {
	apps := Apps{
		lookup:  make(map[string]*App),
		byClass: make(map[string]*AppClass),
	}
	data, err := readFavorites()
	if err != nil {
		log.Println(err)
	}
	list, err := FavoritsUnmarshalJSON(data)
	if err != nil {
		log.Println(err)
	}
	log.Printf("Loaded %d favorites", len(list))
	for _, aClass := range list {
		if aClass.Class != "" {
			aClass.ID = classCnt
			apps.byClass[aClass.Class] = &AppClass{
				ID:          classCnt,
				Class:       aClass.Class,
				Apps:        []*App{},
				Focused:     false,
				Favorite:    true,
				Desktop:     aClass.Desktop,
				Icon:        aClass.Icon,
				LastFocused: "",
			}
			classCnt++
		} else {
			log.Println("Loaded favorite with empty class.")
		}
	}
	return apps
}

type AppList []AppClass

// Delete removes an app from the Apps data structure.
func (a *Apps) Delete(app *App) error {
	// Check if the app exists in the lookup map
	if _, ok := a.lookup[app.Address]; !ok {
		return fmt.Errorf("app with address %s not found", app.Address)
	}

	// Remove the app from the lookup map
	delete(a.lookup, app.Address)

	// Update the class map
	if classInfo, ok := a.byClass[app.Class]; ok {
		// Find and remove the app from the Apps list of the corresponding class
		var newList []*App
		for _, a2 := range classInfo.Apps {
			if a2.Address != app.Address {
				newList = append(newList, a2)
			}
		}

		// Update the Apps list of the class
		a.byClass[app.Class].Apps = newList

		// Update the LastFocused and Favorite status
		if len(newList) == 0 {
			classInfo.LastFocused = ""
			if !classInfo.Favorite {
				// If the class is not a favorite and has no apps, remove it from the class map
				delete(a.byClass, app.Class)
			}
		}
	}

	return nil
}

func (a *Apps) Add(window Client) {
	newApp := &App{
		ID:      AddCnt,
		Color:   activeColor,
		Title:   window.Title,
		Class:   window.Class,
		Address: window.Address,
		Focused: false,
		updated: true,
	}
	a.lookup[window.Address] = newApp
	AddCnt++
	if c, ok := a.byClass[window.Class]; ok {
		c.Apps = append(c.Apps, newApp)
		c.LastFocused = window.Address
	} else {
		desktop, icon := icon.GetIcon(window.Class)
		if desktop != "" {
			desktop = path.Base(desktop)
		}
		a.byClass[window.Class] = &AppClass{
			ID:          classCnt,
			Class:       window.Class,
			Apps:        []*App{newApp},
			Focused:     false,
			Icon:        icon,
			Desktop:     desktop,
			LastFocused: window.Address,
			Favorite:    false,
		}
		classCnt++
	}
}

func (a *Apps) Upsert(window Client) {
	// ignore cases
	if window.Address == "" || window.Class == "" {
		return
	}
	// handle swallowed apps
	if swallowed := window.Swallowing; swallowed != "" {
		if w, ok := a.lookup[swallowed]; ok {
			w.Swallowed = true
		}
		// TODO: when swallowed window is not yet added
	}
	if _, ok := a.lookup[window.Address]; ok {
		a.lookup[window.Address].updated = true
	} else { // add
		a.Add(window)
	}
}

func (a AppList) Len() int           { return len(a) }
func (a AppList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a AppList) Less(i, j int) bool { return a[i].ID < a[j].ID }

func (a AppList) String() string {
	json, err := json.Marshal(a)
	if err != nil {
		return ""
	}
	return string(json)
}

func (a *Apps) Update(c Clients) (alist AppList) {
	// reset state toggles
	for _, a2 := range a.lookup {
		a2.updated = false
		a2.Swallowed = false
		a2.Focused = false
	}
	for _, a2 := range a.byClass {
		a2.Focused = false
	}
	for _, window := range c {
		// ignore special workspaces
		if window.Workspace.ID < 0 {
			continue
		}
		a.Upsert(window)
	}

	// mark focused
	focused := FocusedClient().Address
	if app, ok := a.lookup[focused]; ok {
		app.Focused = true
		a.byClass[app.Class].Focused = true
		a.byClass[app.Class].LastFocused = app.Address
	}

	// delete all not updated and append to list for display
	for _, a2 := range a.lookup {
		if !a2.updated {
			a.Delete(a2)
			continue
		}
	}

	// format into list
	for _, c := range a.byClass {
		alist = append(alist, *c)
	}
	sort.Sort(alist)
	return
}

func FocusedClient() Client {
	var outb, errb bytes.Buffer
	var client Client
	cmd := exec.Command("/usr/bin/hyprctl", "activewindow", "-j")
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Println("running hyprctl activewindow: ", err)
	}
	if err := json.Unmarshal(outb.Bytes(), &client); err != nil {
		log.Println("unmarshal activewindow: ", err)
	}
	return client
}

// Create a temporary struct to hold only the necessary data for marshaling
type favorite struct {
	Desktop string
	Class   string
	Icon    string
}

func FavoritsMarshalJSON(list AppList) ([]byte, error) {
	// Populate the temporary struct
	var tempList []favorite
	for _, appClass := range list {
		// if !appClass.Favorite {
		//	continue
		// }
		tempList = append(tempList, favorite{
			Class:   appClass.Class,
			Desktop: appClass.Desktop,
			Icon:    appClass.Icon,
		})
	}

	// Marshal the temporary struct
	return json.Marshal(tempList)
}

func FavoritsUnmarshalJSON(data []byte) (list AppList, err error) {
	// Create a temporary struct to hold the unmarshaled data
	var tempList []favorite

	// Unmarshal the data into the temporary struct
	if err := json.Unmarshal(data, &tempList); err != nil {
		return list, err
	}

	// Populate the AppList slice with the unmarshaled data (ignoring Apps field)
	for i, tempAppClass := range tempList {
		list = append(list, AppClass{
			Apps:     make([]*App, 0),
			ID:       i,
			Favorite: true,
			Desktop:  tempAppClass.Desktop,
			Class:    tempAppClass.Class,
			Icon:     tempAppClass.Icon,
		})
	}

	return list, nil
}

func SaveFavorites(data []byte) error {
	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}

	// Construct the full path to the config directory and the file
	configDir := filepath.Join(homeDir, ".config", "hypr")
	filePath := filepath.Join(configDir, "hyprshell.json")

	// Create the directory if it doesn't exist
	err = os.MkdirAll(configDir, 0o755)
	if err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Write the data to the file
	err = os.WriteFile(filePath, data, 0o644)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func readFavorites() ([]byte, error) {
	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %w", err)
	}

	// Construct the full path to the config directory and the file
	configDir := filepath.Join(homeDir, ".config", "hypr")
	filePath := filepath.Join(configDir, "hyprshell.json")

	// Check if the file exists
	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("favorites file does not exist")
	}

	// Read the data from the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading from file: %w", err)
	}

	return data, nil
}
