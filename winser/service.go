package winser

import (
	"fmt"
	"log"
	"os"

	"github.com/kardianos/service"
)

var logger service.Logger

type RunServiceFn func(<-chan struct{}, service.Logger, string) error

type program struct {
	exit       chan struct{}
	Configfile string
	FnRun      RunServiceFn
}

func (p *program) Start(s service.Service) error {
	if service.Interactive() {
		logger.Info("Running in terminal.")
	} else {
		logger.Info("Running under service manager.")
	}
	p.exit = make(chan struct{})

	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) run() error {
	// Non blocking routine could be used for all process that are blocking the service start routine
	return p.FnRun(p.exit, logger, p.Configfile)
}

func (p *program) Stop(s service.Service) error {
	logger.Info("Stop the service")
	close(p.exit)
	return nil
}

func HandleAsManagedService(cmd string, cfgFile string, serviceName string, runner RunServiceFn) error {
	svcConfig := &service.Config{
		Name:        fmt.Sprintf("%sService", serviceName),
		DisplayName: fmt.Sprintf("%s Web Service", serviceName),
		Description: fmt.Sprintf("This is the %s Web Service", serviceName),
	}

	prg := &program{
		Configfile: cfgFile,
		FnRun:      runner,
	}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	errs := make(chan error, 5)
	logger, err = s.Logger(errs)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Program called with %v num %d, cmd is %s\n", os.Args[0], len(os.Args), cmd)

	go func() { // Error writer in background
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	if len(cmd) != 0 && cmd != "like" {
		err := service.Control(s, cmd)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		return err
	}
	log.Println("Run the service using Run()")
	err = s.Run()
	if err != nil {
		logger.Error(err)
		return err
	}

	log.Println("Finally HandleAsManagedService is terminated")
	return nil
}
