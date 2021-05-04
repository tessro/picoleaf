package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"gopkg.in/ini.v1"
)

const defaultConfigFile = ".picoleafrc"

var verbose = flag.Bool("v", false, "Verbose")

func usage() {
	fmt.Println("usage: picoleaf [-v] <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println()
	fmt.Println("   on           Turn on Nanoleaf")
	fmt.Println("   off          Turn off Nanoleaf")
	fmt.Println()
	fmt.Println("   effect       Control Nanoleaf effects")
	fmt.Println()
	fmt.Println("   hsl          Set Nanoleaf to the provided HSL")
	fmt.Println("   rgb          Set Nanoleaf to the provided RGB")
	fmt.Println("   temp         Set Nanoleaf to the provided color temperature")
	fmt.Println("   brightness   Set Nanoleaf to the provided brightness")
	fmt.Println()
	os.Exit(1)
}

func main() {
	flag.Parse()

	usr, err := user.Current()
	if err != nil {
		fmt.Printf("error: failed to fetch current user: %v", err)
		os.Exit(1)
	}
	dir := usr.HomeDir
	configFilePath := filepath.Join(dir, defaultConfigFile)

	cfg, err := ini.Load(configFilePath)
	if err != nil {
		fmt.Printf("error: failed to read file: %v", err)
		os.Exit(1)
	}

	client := Client{
		Host:    cfg.Section("").Key("host").String(),
		Token:   cfg.Section("").Key("access_token").String(),
		Verbose: *verbose,
	}

	if *verbose {
		fmt.Printf("Host: %s\n\n", client.Host)
	}

	if flag.NArg() > 0 {
		cmd := flag.Arg(0)
		switch cmd {
		case "brightness":
			doBrightnessCommand(client, flag.Args()[1:])
		case "effect":
			doEffectCommand(client, flag.Args()[1:])
		case "hsl":
			doHSLCommand(client, flag.Args()[1:])
		case "off":
			err = client.Off()
			if err != nil {
				fmt.Printf("error: failed to turn off Nanoleaf: %v", err)
				os.Exit(1)
			}
		case "on":
			err = client.On()
			if err != nil {
				fmt.Printf("error: failed to turn on Nanoleaf: %v", err)
				os.Exit(1)
			}
		case "rgb":
			doRGBCommand(client, flag.Args()[1:])
		case "temp":
			doColorTemperatureCommand(client, flag.Args()[1:])
		default:
			usage()
		}
	} else {
		usage()
	}
}

func doBrightnessCommand(client Client, args []string) {
	if len(args) < 1 {
		fmt.Println("usage: picoleaf brightness <brightness>")
		os.Exit(1)
	}

	brightness, err := strconv.Atoi(args[0])
	if err != nil || brightness < 0 || brightness > 100 {
		fmt.Println("error: temperature must be an integer 0-100")
		os.Exit(1)
	}

	err = client.SetBrightness(brightness)
	if err != nil {
		fmt.Printf("error: failed to set brightness: %v", err)
		os.Exit(1)
	}
}

func doColorTemperatureCommand(client Client, args []string) {
	if len(args) < 1 {
		fmt.Println("usage: picoleaf temp <temperature>")
		os.Exit(1)
	}

	temp, err := strconv.Atoi(args[0])
	if err != nil || temp < 1200 || temp > 6500 {
		fmt.Println("error: temperature must be an integer 1200-6500")
		os.Exit(1)
	}

	err = client.SetColorTemperature(temp)
	if err != nil {
		fmt.Printf("error: failed to set color temperature: %v", err)
		os.Exit(1)
	}
}

func doEffectCommand(client Client, args []string) {
	usage := func() {
		fmt.Println("usage: picoleaf effect list")
		fmt.Println("       picoleaf effect select <name>")
		os.Exit(1)
	}

	if len(args) < 1 {
		usage()
	}

	command := args[0]
	switch command {
	case "list":
		list, err := client.ListEffects()
		if err != nil {
			fmt.Printf("error: failed retrieve effects list: %v", err)
			os.Exit(1)
		}
		for _, name := range list {
			fmt.Println(name)
		}
	case "select":
		if len(args) != 2 {
			fmt.Println("usage: picoleaf effect select <name>")
			os.Exit(1)
		}

		name := args[1]
		err := client.SelectEffect(name)
		if err != nil {
			fmt.Printf("error: failed to select effect: %v", err)
			os.Exit(1)
		}
	default:
		usage()
	}
}

func doHSLCommand(client Client, args []string) {
	if len(args) != 3 {
		fmt.Println("usage: picoleaf hsl <hue> <saturation> <lightness>")
		os.Exit(1)
	}

	hue, err := strconv.Atoi(args[0])
	if err != nil || hue < 0 || hue > 360 {
		fmt.Println("error: hue must be an integer 0-100")
		os.Exit(1)
	}

	sat, err := strconv.Atoi(args[1])
	if err != nil || sat < 0 || sat > 100 {
		fmt.Println("error: saturation must be an integer 0-360")
		os.Exit(1)
	}

	lightness, err := strconv.Atoi(args[2])
	if err != nil || lightness < 0 || lightness > 100 {
		fmt.Println("error: lightness must be an integer 0-100")
		os.Exit(1)
	}

	err = client.SetHSL(hue, sat, lightness)
	if err != nil {
		fmt.Printf("error: failed to set HSL: %v", err)
		os.Exit(1)
	}
}

func doRGBCommand(client Client, args []string) {
	if len(args) != 3 {
		fmt.Println("usage: picoleaf rgb <red> <green> <blue>")
		os.Exit(1)
	}

	red, err := strconv.Atoi(args[0])
	if err != nil || red < 0 || red > 255 {
		fmt.Println("error: red must be an integer 0-255")
		os.Exit(1)
	}

	green, err := strconv.Atoi(args[1])
	if err != nil || green < 0 || green > 255 {
		fmt.Println("error: green must be an integer 0-255")
		os.Exit(1)
	}

	blue, err := strconv.Atoi(args[2])
	if err != nil || blue < 0 || blue > 255 {
		fmt.Println("error: blue must be an integer 0-255")
		os.Exit(1)
	}

	err = client.SetRGB(red, green, blue)
	if err != nil {
		fmt.Printf("error: failed to set RGB: %v", err)
		os.Exit(1)
	}
}
