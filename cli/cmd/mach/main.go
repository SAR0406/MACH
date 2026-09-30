package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/SAR0406/MACH/internal/api"
	"github.com/SAR0406/MACH/internal/config"
	"github.com/SAR0406/MACH/internal/identity"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]
	var err error

	switch cmd {
	case "init":
		err = runInit(args)
	case "status":
		err = runStatus(args)
	case "expose":
		err = runExpose(args)
	case "ssh":
		err = runSSH(args)
	case "logs":
		err = runLogs(args)
	case "stop":
		err = runStop(args)
	case "doctor":
		err = runDoctor(args)
	default:
		usage()
		err = fmt.Errorf("unknown command: %s", cmd)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	name := fs.String("name", hostnameOr("mach-node"), "logical device name")
	controller := fs.String("controller", envOr("MACH_CONTROLLER_URL", "http://127.0.0.1:8080"), "controller URL")
	if err := fs.Parse(args); err != nil {
		return err
	}

	pub, priv, err := identity.GenerateKeypair()
	if err != nil {
		return err
	}
	id, err := identity.DeviceIDFromPublicKey(pub)
	if err != nil {
		return err
	}

	cfg := config.DeviceConfig{
		Version:       "v1",
		DeviceID:      id,
		DeviceName:    *name,
		PublicKey:     pub,
		PrivateKey:    priv,
		ControllerURL: *controller,
	}

	client := api.NewClient(*controller, "")
	ch, err := client.RegisterChallenge(api.RegisterChallengeRequest{DeviceID: id, DeviceName: *name, PublicKey: pub})
	if err != nil {
		return err
	}
	sig, err := identity.Sign(priv, []byte(ch.Challenge))
	if err != nil {
		return err
	}
	reg, err := client.RegisterComplete(api.RegisterCompleteRequest{DeviceID: id, Signature: sig})
	if err != nil {
		return err
	}
	cfg.Token = reg.Token
	if err := config.SaveDeviceConfig(cfg); err != nil {
		return err
	}

	fmt.Printf("✓ Device claimed\n")
	fmt.Printf("✓ Private identity created\n")
	fmt.Printf("✓ Controller linked: %s\n", *controller)
	fmt.Printf("Device ID: %s\n", id)
	return nil
}

func runStatus(args []string) error {
	if len(args) != 0 {
		return errors.New("status takes no arguments")
	}
	cfg, client, err := loadClient()
	if err != nil {
		return err
	}
	st, err := client.Status()
	if err != nil {
		return err
	}
	fmt.Printf("Device: %s (%s)\n", st.DeviceName, cfg.DeviceID)
	fmt.Printf("Online: %v\n", st.Online)
	fmt.Printf("Path: %s\n", st.NetworkPath)
	fmt.Printf("Last heartbeat: %s\n", st.LastHeartbeat.Format(time.RFC3339))
	fmt.Printf("CPU: %d cores\nRAM: %d MiB\n", st.Capabilities.CPUCores, st.Capabilities.MemoryMiB)
	if len(st.Exposures) == 0 {
		fmt.Println("Exposures: none")
		return nil
	}
	fmt.Println("Exposures:")
	for _, ex := range st.Exposures {
		fmt.Printf("- %s port=%d path=%s\n  %s\n  %s\n", ex.ID, ex.Port, ex.Path, ex.PublicURL, ex.RelayURL)
	}
	return nil
}

func runExpose(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: mach expose <port>")
	}
	port, err := parsePort(args[0])
	if err != nil {
		return err
	}
	if !localPortOpen(port) {
		return fmt.Errorf("service not detected on 127.0.0.1:%d", port)
	}
	_, client, err := loadClient()
	if err != nil {
		return err
	}
	ex, err := client.Expose(port)
	if err != nil {
		return err
	}
	fmt.Printf("✓ Service detected\n")
	fmt.Printf("MACH URL: %s\n", ex.PublicURL)
	fmt.Printf("Relay URL: %s\n", ex.RelayURL)
	fmt.Printf("Path: %s\n", ex.Path)
	return nil
}

func runSSH(args []string) error {
	if len(args) != 0 {
		return errors.New("ssh takes no arguments")
	}
	_, client, err := loadClient()
	if err != nil {
		return err
	}
	st, err := client.Status()
	if err != nil {
		return err
	}
	target := st.DeviceID + ".mach.dev"
	if len(st.Exposures) > 0 {
		target = st.Exposures[0].ID + ".mach.dev"
	}
	fmt.Printf("ssh mach@%s\n", target)
	fmt.Println("(SSH transport is relay-capable in MVP and controlled by the agent/control plane.)")
	return nil
}

func runLogs(args []string) error {
	fs := flag.NewFlagSet("logs", flag.ContinueOnError)
	limit := fs.Int("n", 30, "number of log lines")
	if err := fs.Parse(args); err != nil {
		return err
	}
	_, client, err := loadClient()
	if err != nil {
		return err
	}
	logs, err := client.Logs(*limit)
	if err != nil {
		return err
	}
	for _, e := range logs.Entries {
		fmt.Printf("%s %-15s %s\n", e.Timestamp.Format(time.RFC3339), e.Action, e.Message)
	}
	return nil
}

func runStop(args []string) error {
	exposureID := ""
	if len(args) > 1 {
		return errors.New("usage: mach stop [exposure-id]")
	}
	if len(args) == 1 {
		exposureID = args[0]
	}
	_, client, err := loadClient()
	if err != nil {
		return err
	}
	if err := client.Stop(exposureID); err != nil {
		return err
	}
	fmt.Println("✓ Exposure stopped")
	return nil
}

func runDoctor(args []string) error {
	if len(args) != 0 {
		return errors.New("doctor takes no arguments")
	}
	cfg, client, err := loadClient()
	if err != nil {
		return err
	}
	st, err := client.Status()
	if err != nil {
		return err
	}
	fmt.Println("MACH DIAGNOSTICS")
	printCheck("Identity", cfg.DeviceID != "")
	printCheck("Controller", cfg.ControllerURL != "")
	printCheck("Agent heartbeat", st.Online)
	printCheck("Direct path", st.NetworkPath == "DIRECT")
	printCheck("Relay fallback", st.NetworkPath == "RELAY" || st.NetworkPath == "DIRECT")
	printCheck("TLS endpoint assigned", len(st.Exposures) > 0)
	fmt.Println()
	if st.NetworkPath == "RELAY" {
		fmt.Println("Suggested action:")
		fmt.Println("Relay mode is active. No configuration required.")
	}
	return nil
}

func loadClient() (*config.DeviceConfig, *api.Client, error) {
	cfg, err := config.LoadDeviceConfig()
	if err != nil {
		return nil, nil, err
	}
	if cfg.Token == "" {
		return nil, nil, errors.New("device is not registered; run mach init")
	}
	return cfg, api.NewClient(cfg.ControllerURL, cfg.Token), nil
}

func localPortOpen(port int) bool {
	c, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 800*time.Millisecond)
	if err != nil {
		return false
	}
	_ = c.Close()
	return true
}

func parsePort(v string) (int, error) {
	p := 0
	_, err := fmt.Sscanf(v, "%d", &p)
	if err != nil || p <= 0 || p > 65535 {
		return 0, fmt.Errorf("invalid port: %s", v)
	}
	return p, nil
}

func printCheck(name string, ok bool) {
	mark := "✗"
	if ok {
		mark = "✓"
	}
	fmt.Printf("%-22s %s\n", name, mark)
}

func usage() {
	fmt.Println("mach <command>")
	fmt.Println("commands: init status expose ssh logs stop doctor")
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func hostnameOr(def string) string {
	h, err := os.Hostname()
	if err != nil || strings.TrimSpace(h) == "" {
		return def
	}
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(strings.ToLower(h), " ", "-")
	}
	return h
}
