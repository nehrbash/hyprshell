package icon

// Icon Naming Specification
// https://specifications.freedesktop.org/icon-naming-spec/icon-naming-spec-latest.html

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"unicode"
)

type iconInfo struct {
	icon    string
	desktop string
}

var (
	iconMap     map[string]iconInfo
	appDirs     []string
	defaultIcon string
	iconTheme   string
)

func init() {
	iconMap = make(map[string]iconInfo)
	// add any manual hard fixes of class name here.
	var err error
	appDirs = getAppDirs()
	iconTheme, err = getGTKIconThemeName()
	if err != nil {
		panic(err)
	}
	defaultIcon, err = LookupSVGIconPath("default-application")
	if err != nil {
		panic(err)
	}
}

func GetIcon(className string) (string, string) {
	if iconUrl, ok := iconMap[className]; ok {
		return iconUrl.desktop, iconUrl.icon
	}
	var iconPath string
	var guesses []string
	desktopPath, iconName, err := getDesktopIconName(className)
	if err != nil {
		guesses = append(guesses, iconName)
	}
	guesses = append(guesses, className, strings.ToLower(className))
	firstW := getFirstWord(className)
	if firstW != "" {
		guesses = append(guesses, firstW, strings.ToLower(firstW))
	}

	for _, guess := range guesses {
		iconPath, err = LookupSVGIconPath(guess)
		if err == nil {
			break // found a good path
		}
	}
	if iconPath != "" {
		iconMap[className] = iconInfo{icon: iconPath, desktop: desktopPath}
		return desktopPath, iconPath
	}
	// else use default
	iconMap[className] = iconInfo{icon: defaultIcon, desktop: desktopPath}
	return desktopPath, defaultIcon
}

func isSeparator(r rune) bool {
	return unicode.IsSpace(r) || r == '_' || r == '-'
}

func getFirstWord(s string) string {
	words := strings.FieldsFunc(s, isSeparator)
	if len(words) > 0 {
		return words[0]
	}
	return ""
}

func getDesktopIconName(appName string) (desktopPath string, icon string, err error) {

	for _, d := range appDirs {
		path := filepath.Join(d, fmt.Sprintf("%s.desktop", appName))
		if pathExists(path) {
			desktopPath = path
		} else if pathExists(strings.ToLower(path)) {
			desktopPath = strings.ToLower(path)
		}
	}
	/* Some apps' app_id varies from their .desktop file name, e.g. 'gimp-2.9.9' or 'pamac-manager'.
	   Let's try to find a matching .desktop file name */
	if !strings.HasPrefix(appName, "/") && desktopPath == "" { // skip icon paths given instead of names
		desktopPath = searchDesktopDirs(appName)
	}

	if desktopPath != "" {
		lines, err := loadTextFile(desktopPath)
		if err != nil {
			return desktopPath, "", err
		}
		for _, line := range lines {
			if strings.HasPrefix(strings.ToUpper(line), "ICON") {
				parts := strings.Split(line, "=")
				if len(parts) >= 2 {
					return desktopPath, parts[1], nil
				}
			}
		}
	}
	return desktopPath, "", errors.New("couldn't find the icon")
}

func getAppDirs() []string {
	var dirs []string
	xdgDataDirs := ""

	home := os.Getenv("HOME")
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if os.Getenv("XDG_DATA_DIRS") != "" {
		xdgDataDirs = os.Getenv("XDG_DATA_DIRS")
	} else {
		xdgDataDirs = "/usr/local/share/:/usr/share/"
	}
	if xdgDataHome != "" {
		dirs = append(dirs, filepath.Join(xdgDataHome, "applications"))
	} else if home != "" {
		dirs = append(dirs, filepath.Join(home, ".local/share/applications"))
	}
	for _, d := range strings.Split(xdgDataDirs, ":") {
		dirs = append(dirs, filepath.Join(d, "applications"))
	}
	flatpakDirs := []string{filepath.Join(home, ".local/share/flatpak/exports/share/applications"),
		"/var/lib/flatpak/exports/share/applications"}

	for _, d := range flatpakDirs {
		if !isIn(dirs, d) {
			dirs = append(dirs, d)
		}
	}
	return dirs
}

func isIn(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func pathExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func searchDesktopDirs(badAppID string) string {
	b4Hyphen := strings.Split(badAppID, "-")[0]
	for _, d := range appDirs {
		items, _ := os.ReadDir(d)
		for _, item := range items {
			if strings.Contains(item.Name(), b4Hyphen) {
				//Let's check items starting from 'org.' first
				if strings.Count(item.Name(), ".") > 1 {
					return filepath.Join(d, item.Name())
				}
			}
		}
	}
	return ""
}

func loadTextFile(path string) ([]string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(bytes), "\n")
	var output []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			output = append(output, line)
		}
	}
	return output, nil
}

func LookupSVGIconPath(iconName string) (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	directories := []string{
		filepath.Join(user.HomeDir, ".icons", iconTheme),
		filepath.Join("/usr/share/icons", iconTheme),
	}

	for _, directory := range directories {
		iconPath, err := searchSVGIconPath(directory, iconName)
		if err == nil {
			return iconPath, nil
		}
	}

	return "", fmt.Errorf("icon not found")
}

func searchSVGIconPath(baseDir, iconName string) (string, error) {
	iconSizeDir := filepath.Join(baseDir, "scalable")
	iconPath := filepath.Join(iconSizeDir, "apps", iconName+".svg")

	if _, err := os.Stat(iconPath); err == nil {
		return iconPath, nil
	}

	files, err := ioutil.ReadDir(iconSizeDir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) == ".svg" && strings.TrimSuffix(file.Name(), ".svg") == iconName {
			return filepath.Join(iconSizeDir, file.Name()), nil
		}
	}

	return "", fmt.Errorf("icon not found in %s", iconSizeDir)
}

func getGTKIconThemeName() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}

	configPaths := []string{
		filepath.Join(user.HomeDir, ".config", "gtk-3.0", "settings.ini"),
		"/etc/gtk-3.0/settings.ini",
	}

	for _, configPath := range configPaths {
		iconThemeName, err := extractIconThemeName(configPath)
		if err == nil {
			return iconThemeName, nil
		}
	}

	return "", fmt.Errorf("GTK Icon Theme Name not found")
}

func extractIconThemeName(configPath string) (string, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "gtk-icon-theme-name=") {
			return strings.TrimPrefix(line, "gtk-icon-theme-name="), nil
		}
	}

	return "", fmt.Errorf("GTK Icon Theme Name not found in %s", configPath)
}
