package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sunshineplan/metadata"
	"github.com/sunshineplan/password"
	"github.com/sunshineplan/service"
	"github.com/sunshineplan/utils"
	"github.com/sunshineplan/utils/flags"
	"github.com/sunshineplan/utils/httpsvr"
)

var self string
var meta metadata.Server
var priv *rsa.PrivateKey

var server = httpsvr.New()
var svc = service.Service{
	Name:     "MyTasks",
	Desc:     "Instance to serve My Tasks",
	Exec:     run,
	TestExec: test,
	Options: service.Options{
		Dependencies:       []string{"Wants=network-online.target", "After=network.target"},
		Environment:        map[string]string{"GIN_MODE": "release"},
		RemoveBeforeUpdate: []string{"dist/assets"},
		ExcludeFiles:       []string{"scripts/mytasks.conf"},
	},
}

var (
	joinPath = filepath.Join
	dir      = filepath.Dir
)

var (
	maxRetry  = flag.Int("retry", 5, "Max number of retries on wrong password")
	universal = flag.Bool("universal", false, "Use Universal account id or not")
	pemPath   = flag.String("pem", "", "PEM File Path")
	logPath   = flag.String("log", "", "Log Path")
	// logPath = flag.String("log", joinPath(dir(self), "access.log"), "Log Path")
)

func main() {
	var err error
	self, err = os.Executable()
	if err != nil {
		log.Fatalln("Failed to get self path:", err)
	}

	flag.StringVar(&meta.Addr, "server", "", "Metadata Server Address")
	flag.StringVar(&meta.Header, "header", "", "Verify Header Header Name")
	flag.StringVar(&meta.Value, "value", "", "Verify Header Value")
	flag.StringVar(&server.Unix, "unix", "", "UNIX-domain Socket")
	flag.StringVar(&server.Host, "host", "0.0.0.0", "Server Host")
	flag.StringVar(&server.Port, "port", "12345", "Server Port")
	flag.StringVar(&svc.Options.UpdateURL, "update", "", "Update URL")
	flags.SetConfigFile(joinPath(dir(self), "config.ini"))
	flags.Parse()

	password.SetMaxAttempts(*maxRetry)
	if *pemPath != "" {
		b, err := os.ReadFile(*pemPath)
		if err != nil {
			log.Fatal(err)
		}
		block, _ := pem.Decode(b)
		if block == nil {
			log.Fatal("no PEM data is found")
		}
		priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			log.Fatal(err)
		}
	}

	if service.IsWindowsService() {
		svc.Run(false)
		return
	}

	switch flag.NArg() {
	case 0:
		run()
	case 1:
		cmd := flag.Arg(0)
		var ok bool
		if ok, err = svc.Command(cmd); !ok {
			if cmd == "add" || cmd == "delete" {
				log.Fatalf("%s need two arguments", cmd)
			} else {
				log.Fatalln("Unknown argument:", cmd)
			}
		}
	case 2:
		switch flag.Arg(0) {
		case "add":
			addUser(flag.Arg(1))
		case "delete":
			if utils.Confirm(fmt.Sprintf("Do you want to delete user %s?", flag.Arg(1)), 3) {
				deleteUser(flag.Arg(1))
			}
		default:
			log.Fatalln("Unknown arguments:", strings.Join(flag.Args(), " "))
		}
	default:
		log.Fatalln("Unknown arguments:", strings.Join(flag.Args(), " "))
	}
	if err != nil {
		log.Fatalf("Failed to %s: %v", flag.Arg(0), err)
	}
}
