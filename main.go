package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/crc-org/vfkit/pkg/config"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	base := filepath.Join(home, ".vmlab", "debian")
	if err := os.MkdirAll(base, 0o755); err != nil {
		log.Fatal(err)
	}

	efiVars := filepath.Join(base, "efi-vars")
	_, statErr := os.Stat(efiVars)
	createEFI := os.IsNotExist(statErr)

	vm := config.NewVirtualMachine(2, 2048,
		config.NewEFIBootloader(efiVars, createEFI))

	disk, err := config.VirtioBlkNew(filepath.Join(home, ".vmlab", "images", "debian.raw"))
	if err != nil {
		log.Fatal(err)
	}
	netdev, err := config.VirtioNetNew("") // empty MAC -> random
	if err != nil {
		log.Fatal(err)
	}
	serial, err := config.VirtioSerialNewStdio()
	if err != nil {
		log.Fatal(err)
	}
	if err := vm.AddDevices(disk, netdev, serial); err != nil {
		log.Fatal(err)
	}

	cmd, err := vm.Cmd("vfkit")
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	<-ctx.Done()
	_ = cmd.Process.Signal(syscall.SIGTERM)
	_ = cmd.Wait()
}
