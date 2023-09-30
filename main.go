package main


import (
	"bufio"
	"fmt"
	"os"
	"path"
)

// Builds a config and checks if it exists
func GetConfigPath(base string, hostname string) (string, error) {
	configPath := path.Join(base, hostname, "host.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", err
	}
	return configPath, nil
}

func run(base string, hostname string) {
	configPath, err := GetConfigPath(base, hostname)
	if err != nil {
		fmt.Println(err)
		fmt.Errorf("Config file %s not found", configPath)
		return
	}

}

func main() {
	reader := bufio.NewReader(os.Stdin)
	hostname, _ := reader.ReadString('\n')
	fmt.Printf("Installing for: %s", hostname)
	base := os.Getenv("HOME") + "/dotfiles"
	run(base, hostname)
}
