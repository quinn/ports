package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type mappings map[string]map[string]int

func main() {
	namespaceFlag := flag.String("namespace", defaultNamespace(), "namespace for service mappings")
	flag.Parse()

	if flag.NArg() == 0 {
		exitf("missing required service argument")
	}
	if flag.NArg() > 1 {
		exitf("unexpected positional arguments: %s", strings.Join(flag.Args()[1:], " "))
	}

	namespace := strings.TrimSpace(*namespaceFlag)
	if namespace == "" {
		namespace = defaultNamespace()
	}

	service := strings.TrimSpace(flag.Arg(0))
	if service == "" {
		exitf("service argument cannot be empty")
	}

	configPath, err := configFilePath()
	if err != nil {
		exitErr(err)
	}

	data, err := loadMappings(configPath)
	if err != nil {
		exitErr(err)
	}

	if data[namespace] == nil {
		data[namespace] = make(map[string]int)
	}

	if port, ok := data[namespace][service]; ok {
		fmt.Println(port)
		return
	}

	port, err := findAvailablePort(data)
	if err != nil {
		exitErr(err)
	}

	data[namespace][service] = port
	if err := saveMappings(configPath, data); err != nil {
		exitErr(err)
	}

	fmt.Println(port)
}

func defaultNamespace() string {
	wd, err := os.Getwd()
	if err != nil {
		return "default"
	}

	base := filepath.Base(wd)
	if base == "" || base == "." || base == string(filepath.Separator) {
		return "default"
	}

	return base
}

func configFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine home directory: %w", err)
	}
	return filepath.Join(home, ".ports.json"), nil
}

func loadMappings(path string) (mappings, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(mappings), nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	if len(b) == 0 {
		return make(mappings), nil
	}

	var out mappings
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if out == nil {
		out = make(mappings)
	}

	return out, nil
}

func saveMappings(path string, data mappings) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal mappings: %w", err)
	}
	b = append(b, '\n')

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, b, 0o600); err != nil {
		return fmt.Errorf("write temp file %s: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}

	return nil
}

func findAvailablePort(data mappings) (int, error) {
	usedInJson := make(map[int]bool)
	for _, services := range data {
		for _, port := range services {
			usedInJson[port] = true
		}
	}

	for port := 3000; port <= 7999; port++ {
		if usedInJson[port] {
			continue
		}

		if isPortFree(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports found in range 3000-7999")
}

func isPortFree(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func exitf(format string, args ...any) {
	exitErr(fmt.Errorf(format, args...))
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
