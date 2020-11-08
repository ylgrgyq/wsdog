package main

import (
	"github.com/jessevdk/go-flags"
	"os"
)

type ApplicationOptions struct {
	ListenPort  uint16 `short:"l" long:"listen" description:"listen on port"`
	ConnectUrl  string `short:"c" long:"connect" description:"connect to a WebSocket server"`
	EnableDebug bool   `long:"debug" description:"enable debug log"`
	NoColor     bool   `long:"no-color" description:"Run without color"`
	ShowPingPong bool   `short:"P" long:"show-ping-pong" description:"print a notification when a ping or pong is received"`
	Subprotocol  string `short:"s" long:"subprotocol" description:"optional subprotocol (default: )"`
}

type ListenOnPortOptions struct {

}

type ConnectOptions struct {
	Origin         string            `short:"o" long:"origin" description:"optional origin"`
	ExecuteCommand string            `short:"x" long:"execute" description:"execute command after connecting"`
	Wait           int64             `short:"w" long:"wait" default:"2" description:" wait given seconds after executing command"`
	Host           string            `long:"host" description:"optional host"`
	NoTlsCheck     bool              `short:"n" long:"no-check" description:"Do not check for unauthorized certificates"`
	Headers        map[string]string `short:"H" long:"header" description:"Set an HTTP header <header:value>. Repeat to set multiple like -H header1:value1 -H header2:value2."`
	Auth           string            `long:"auth" description:"Add basic HTTP authentication header <username:password>."`
	//Ca             string            `long:"ca" description:"Specify a Certificate Authority"`
	//Cert           string            `long:"cert" description:"Specify a Client SSL Certificate"`
	//Key            string            `long:"key" description:"Specify a Client SSL Certificate's key"`
	//Passphrase     string            `long:"passphrase" description:"Specify a Client SSL Certificate Key's passphrase. If you don't provide a value, it will be prompted for."`
	EnableSlash    bool              `long:"slash" description:"Enable slash commands for control frames (/ping, /pong, /close [code [, reason]])"`
}

type CommandLineOptions struct {
	ApplicationOptions
	ListenOnPortOptions
	ConnectOptions
}

func parseCommandLineArguments() CommandLineOptions {
	var appOpts ApplicationOptions
	var listenOptions ListenOnPortOptions
	var connectOptions ConnectOptions
	parser := flags.NewParser(&appOpts, flags.Default)

	_, _ = parser.AddGroup(
		"Listen On Port Options",
		"Listen On Port Options",
		&listenOptions)
	_, _ = parser.AddGroup(
		"Connect To A WebSocket Server Options",
		"Connect To A WebSocket Server Options",
		&connectOptions)
	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case *flags.Error:
			if flagsErr.Type == flags.ErrHelp {
				os.Exit(0)
			}

			parser.WriteHelp(os.Stderr)
			os.Exit(1)
		default:
			parser.WriteHelp(os.Stderr)
			os.Exit(1)
		}
	}

	if appOpts.ConnectUrl == "" && appOpts.ListenPort == 0 {
		parser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	if appOpts.NoColor {
		SetLogger(noColorLogger)
	}

	if appOpts.EnableDebug {
		defaultLogger.EnableDebug()
		noColorLogger.EnableDebug()
	}

	return CommandLineOptions{appOpts, listenOptions, connectOptions}
}

func main() {
	var cliOpts = parseCommandLineArguments()

	if cliOpts.ConnectUrl != "" {
		runAsClient(cliOpts.ConnectUrl, cliOpts)
	} else {
		runAsServer(cliOpts.ListenPort, cliOpts)
	}
}
