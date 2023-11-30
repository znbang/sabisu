package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kardianos/service"
)

var logger service.Logger

const Usage = `Windows Service Wrapper.

Usage:
  %v <command> <configuration file>

<command> can be one of:
  -c  --console run as console application
  -t  --start   start the service
  -p  --stop    stop the service
  -e  --restart restart the service
  -i  --install install the service
  -r  --remove  uninstall the service
  -?  --help    print this help message

<configuration file> is the path to config file`

const ActionInstall = "install"
const ActionUninstall = "uninstall"
const ActionStart = "start"
const ActionStop = "stop"
const ActionRestart = "restart"
const ActionHelp = "help"

func getAction() string {
	action := os.Args[1]

	switch action {
	case "-c", "--console", "-s", "--service":
		action = ""
	case "-t", "-start":
		action = ActionStart
	case "-p", "--stop":
		action = ActionStop
	case "-e", "--restart":
		action = ActionRestart
	case "-i", "--install":
		action = ActionInstall
	case "-r", "--remove":
		action = ActionUninstall
	case "-?", "-h", "--help":
		action = ActionHelp
	default:
		action = ActionHelp
	}

	return action
}

func showUsage() {
	exe := os.Args[0]
	ext := filepath.Ext(exe)
	if ext == ".exe" {
		exe = strings.TrimSuffix(exe, ext)
	}
	fmt.Println(fmt.Sprintf(Usage, exe))
	os.Exit(1)
}

func getExeDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Dir(exePath), nil
}

func getAbsPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	if service.Interactive() {
		return filepath.Abs(path)
	} else {
		exeDir, err := getExeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(exeDir, path), nil
	}
}

func main() {
	if len(os.Args) != 3 {
		showUsage()
	}

	// parsing arguments
	action := getAction()
	if action == ActionHelp {
		showUsage()
	}

	// loading config
	configAbsPath, err := getAbsPath(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}

	config, err := loadConfig(configAbsPath)
	if err != nil {
		log.Fatal(err)
	}

	svcConfig := &service.Config{
		Name:        config.Service.Name,
		DisplayName: config.Service.DisplayName,
		Description: config.Service.Description,
		Arguments:   []string{"-s", configAbsPath},
		Option: service.KeyValue{
			"Interactive": config.Service.Interactive,
		},
	}

	switch config.Service.StartType {
	case StartTypeManual:
		svcConfig.Option[service.StartType] = service.ServiceStartManual
	case StartTypeDisabled:
		svcConfig.Option[service.StartType] = service.ServiceStartDisabled
	case StartTypeAuto:
		svcConfig.Option[service.StartType] = service.ServiceStartAutomatic
	default:
		svcConfig.Option[service.StartType] = service.ServiceStartAutomatic
	}

	prg := &program{}
	prg.exeDir, err = getExeDir()
	if err != nil {
		log.Fatal(err)
	}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	prg.config = config
	prg.service = s

	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	if action != "" {
		if err := service.Control(s, action); err != nil {
			log.Fatal(err)
		}
		return
	}

	err = s.Run()
	if err != nil {
		_ = logger.Error(err)
	}
}
