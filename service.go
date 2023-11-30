package main

import (
	"bufio"
	"context"
	"github.com/kardianos/service"
	"github.com/znbang/logtate"
	"golang.org/x/sys/windows/registry"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"
)

var _ service.Interface = (*program)(nil)
var _ service.Shutdowner = (*program)(nil)

type program struct {
	config  *Config
	service service.Service
	cmd     *exec.Cmd
	exeDir  string
}

func (p *program) Start(s service.Service) error {
	// log to file when running as service
	if !service.Interactive() {
		logAbsPath, err := getAbsPath(p.config.Log.Path)
		if err != nil {
			log.Fatal(err)
		}

		log.SetOutput(logtate.New(logtate.Option{
			Path:      logAbsPath,
			MaxBackup: p.config.Log.MaxBackup,
			MaxSize:   p.config.Log.MaxSize,
		}))
	}

	log.Println("starting service")

	go p.run()
	return nil
}

func isJavaExecutable(cmd string) bool {
	return cmd == "java" || cmd == "java.exe" || cmd == "javaw" || cmd == "javaw.exe"
}

func getJavaPath(cmd string) (string, error) {
	// get current version
	regPath := "SOFTWARE\\JavaSoft\\JDK"

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, regPath, registry.READ)
	if err != nil {
		return "", err
	}
	defer k.Close()

	currentVersion, _, err := k.GetStringValue("CurrentVersion")
	if err != nil {
		return "", err
	}

	// get java home
	jdkPath := regPath + "\\" + currentVersion

	jdkKey, err := registry.OpenKey(registry.LOCAL_MACHINE, jdkPath, registry.READ)
	if err != nil {
		return "", err
	}
	defer jdkKey.Close()

	javaHome, _, err := jdkKey.GetStringValue("JavaHome")
	if err != nil {
		return "", err
	}

	return filepath.Join(javaHome, "bin", cmd), nil
}

func getCmdPath(cmd string) string {
	if filepath.IsAbs(cmd) {
		return cmd
	}

	if isJavaExecutable(cmd) {
		if javaPath, err := getJavaPath(cmd); err == nil {
			return javaPath
		}
	}

	return cmd
}

func (p *program) runCommand() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	p.cmd = exec.CommandContext(ctx, getCmdPath(p.config.Exec.Command), p.config.Exec.Args...)
	p.cmd.Env = append(os.Environ(), p.config.Exec.Envs...)
	p.cmd.Dir = p.exeDir

	if stdoutPipe, err := p.cmd.StdoutPipe(); err != nil {
		log.Println("get stdout pipe failed:", err)
		return err
	} else {
		go func() {
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				log.Println(scanner.Text())
			}
		}()
	}

	if stderrPipe, err := p.cmd.StderrPipe(); err != nil {
		log.Println("get stderr pipe failed:", err)
		return err
	} else {
		go func() {
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				log.Println(scanner.Text())
			}
		}()
	}

	log.Println("exec:", p.cmd.Path)

	if err := p.cmd.Run(); err != nil {
		log.Println("exec failed:", err)
		return err
	}

	return nil
}

func (p *program) run() {
	defer func() {
		if service.Interactive() {
			p.Stop(p.service)
		} else {
			p.service.Stop()
		}
	}()

	p.runCommand()

	if p.config.Service.ExecRetry {
		for retryCount := 0; retryCount < p.config.Service.ExecMaxRetry || p.config.Service.ExecMaxRetry == 0; retryCount++ {
			time.Sleep(time.Second)
			log.Printf("retry %d...\n", retryCount+1)
			p.runCommand()
		}
	}
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	log.Println("stopping service")
	if p.cmd.Process != nil {
		p.cmd.Process.Kill()
	}
	if service.Interactive() {
		os.Exit(0)
	}
	return nil
}

func (p *program) Shutdown(s service.Service) error {
	log.Println("shutdown machine")
	return p.Stop(s)
}
