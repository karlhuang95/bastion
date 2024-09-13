package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"

	"github.com/nsf/termbox-go"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

// Cluster represents a cluster item with a name and an SSH command
type Cluster struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

// Config represents the configuration file structure
type Config struct {
	Clusters []Cluster `yaml:"clusters"`
}

var clusters []Cluster
var selectedIndex = 0
var timeoutDuration = 3600 * time.Second // Set a timeout duration of 10 seconds

func main() {
	// Load clusters from configuration file
	err := loadConfig()
	if err != nil {
		fmt.Printf("加载配置文件失败: %v\n", err)
		return
	}

	// Initialize termbox
	err = termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	// Initialize function selector inside the main function
	functionSelector := map[string]func(){
		"executeCommand": executeCommand,
		"drawMenu":       drawMenu,
		"handleTimeout":  handleTimeout,
	}

	// Draw the initial menu
	functionSelector["drawMenu"]()

	// Event loop
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			handleKeyEvent(ev, functionSelector)
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

// loadConfig loads the configuration from the ~/.s/s.yaml file
func loadConfig() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	configPath := filepath.Join(usr.HomeDir, ".s", "s.yaml")

	// Read the YAML configuration file
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件 %s 失败: %v", configPath, err)
	}

	// Unmarshal the YAML data into the Config struct
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	// Load clusters into the global variable
	clusters = config.Clusters
	return nil
}

// handleKeyEvent handles key events and uses the function selector
func handleKeyEvent(ev termbox.Event, functionSelector map[string]func()) {
	switch ev.Key {
	case termbox.KeyArrowUp:
		if selectedIndex > 0 {
			selectedIndex--
		}
		functionSelector["drawMenu"]() // Redraw the menu
	case termbox.KeyArrowDown:
		if selectedIndex < len(clusters)-1 {
			selectedIndex++
		}
		functionSelector["drawMenu"]() // Redraw the menu
	case termbox.KeyEnter:
		functionSelector["executeCommand"]() // Execute selected command
		functionSelector["drawMenu"]()       // Redraw the menu after returning
	case termbox.KeyEsc:
		termbox.Close() // Properly close termbox before exiting
		os.Exit(0)
	}

	// Handle pressing 'w' to exit
	if ev.Ch == 'q' || ev.Ch == 'Q' { // Allow both 'w' and 'W' to exit
		termbox.Close() // Properly close termbox before exiting
		os.Exit(0)
	}
}

// drawMenu draws the list of clusters with the current selection
func drawMenu() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// Draw a header
	drawHeader()

	// Draw the list of clusters
	for i, cluster := range clusters {
		printLine(2, 3+i, cluster.Name, i == selectedIndex)
	}

	termbox.Flush()
}

// drawHeader draws a header with title and instructions
func drawHeader() {
	headerText := "集群切换工具"
	instructions := "使用 ↑ ↓ 键选择，回车执行，Esc 退出，按 'q' 退出"

	// Set title
	printCenteredLine(1, headerText, termbox.ColorCyan, termbox.ColorDefault)

	// Set instructions
	printCenteredLine(2, instructions, termbox.ColorGreen, termbox.ColorDefault)
}

// printLine prints a line with a selection indicator
func printLine(x, y int, text string, selected bool) {
	var fg, bg termbox.Attribute
	if selected {
		fg, bg = termbox.ColorWhite, termbox.ColorBlue // Selected item color
	} else {
		fg, bg = termbox.ColorBlack, termbox.ColorLightGray // Unselected item color
	}
	for i, ch := range text {
		termbox.SetCell(x+i, y, ch, fg, bg)
	}
}

// printCenteredLine prints text centered horizontally at a given row
func printCenteredLine(y int, text string, fg, bg termbox.Attribute) {
	width, _ := termbox.Size()
	x := (width - len(text)) / 2
	for i, ch := range text {
		termbox.SetCell(x+i, y, ch, fg, bg)
	}
}

// executeCommand runs the SSH command for the selected cluster
func executeCommand() {
	termbox.Close() // Close termbox before executing the command

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
	defer cancel()

	// Prepare the command with context
	cmd := exec.CommandContext(ctx, "bash", "-c", clusters[selectedIndex].Command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Run the command and handle potential errors
	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			handleTimeout() // Call timeout handler
		} else {
			fmt.Printf("执行命令失败: %v\n", err)
		}
	} else {
		fmt.Println("SSH 会话结束。即将返回集群切换界面...")
	}

	// Wait for user input before returning to the main menu
	time.Sleep(2 * time.Second) // Optional: Add a small delay for user to read the message
	termbox.Init()              // Reinitialize termbox after command execution
}

// handleTimeout handles the timeout scenario and returns to the menu
func handleTimeout() {
	fmt.Println("执行超时，请检查网络连接。即将返回集群切换界面...")
	time.Sleep(2 * time.Second) // Optional delay for reading the message
	termbox.Init()              // Reinitialize termbox to return to the main menu
}
