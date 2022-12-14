package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aaaasmile/service-winmusic/util"
	"github.com/aaaasmile/service-winmusic/web"
	"github.com/aaaasmile/service-winmusic/web/idl"
	"github.com/aaaasmile/service-winmusic/winser"
	"github.com/kardianos/service"
)

func main() {
	var configfile = flag.String("config", "config.toml", "Configuration file path")
	var ver = flag.Bool("version", false, "Prints current version")
	var cmd = flag.String("cmd", "", `
	where the string is one of this service control commands: 
	start, stop, restart, install, uninstall or like.
	Commands are used to manage the windows service. After the install, usually, you have to configure the "Log On" user before starting the service.
	Only the command "like" starts the application in console but it is different as an empty command.
	An empty command is used to start the application without the windows service stuff.`)
	var serviceName = flag.String("servicename", idl.WebServiceName, fmt.Sprintf("Set the Windows service install name (default %s)", idl.WebServiceName))

	flag.Parse()

	if *ver {
		fmt.Printf("%s, version: %s", idl.Appname, idl.Buildnr)
		os.Exit(0)
	}

	log.Printf("** Start the program: %s vers: %s **\n", os.Args[0], idl.Buildnr)
	if !service.Interactive() {
		// started from windows service, this log file need to be writable from user logon service
		var f *os.File
		var err error

		fmt.Println("Some output will be redirected to a file log")
		f, err = os.OpenFile(util.GetUserLogFile(*serviceName), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		log.SetOutput(f)
		util.UseRelativeRoot = false
		log.Println("->Service is managed, start file logging. App ver nr: ", idl.Buildnr)
	}

	if *cmd == "" && service.Interactive() {
		log.Println("Start the service as console process")
		if err := web.RunService(nil, nil, *configfile); err != nil {
			log.Fatal("Error: ", err)
		}
	} else {
		if err := winser.HandleAsManagedService(*cmd, *configfile, *serviceName, web.RunService); err == nil {
			log.Printf("Command '%s' executed with success.", *cmd)
		}
	}
	return

}
