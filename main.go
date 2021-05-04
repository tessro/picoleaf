package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"gopkg.in/ini.v1"
)

const defaultConfigFile = ".picoleafrc"

var verbose = flag.Bool("v", false, "Verbose")

// Client is a Nanoleaf REST API client.
type Client struct {
	Host  string
	Token string

	client http.Client
}

// Get performs a GET request.
func (c Client) Get(path string) string {
	if *verbose {
		fmt.Println("\nGET", path)
	}

	url := c.Endpoint(path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	if *verbose {
		fmt.Println("<===", string(body))
	}
	return string(body)
}

// Put performs a PUT request.
func (c Client) Put(path string, body []byte) {
	if *verbose {
		fmt.Println("PUT", path)
		fmt.Println("===>", string(body))
	}

	url := c.Endpoint(path)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	res, err := c.client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}
}

// Endpoint returns the full URL for an API endpoint.
func (c Client) Endpoint(path string) string {
	return fmt.Sprintf("http://%s/api/v1/%s/%s", c.Host, c.Token, path)
}

// ListEffects returns an array of effect names.
func (c Client) ListEffects() ([]string, error) {
	body := c.Get("effects/effectsList")
	var list []string
	err := json.Unmarshal([]byte(body), &list)
	return list, err
}

// SelectEffect activates the specified effect.
func (c Client) SelectEffect(name string) error {
	req := EffectsSelectRequest{
		Select: name,
	}
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	c.Put("effects/select", bytes)
	return nil
}

// SetHSL sets the Nanoleaf's hue, saturation, and lightness (brightness).
func (c Client) SetHSL(hue int, sat int, lightness int) error {
	state := State{
		Brightness: &BrightnessProperty{lightness, 0},
		Hue:        &HueProperty{hue},
		Saturation: &SaturationProperty{sat},
	}

	bytes, err := json.Marshal(state)
	if err != nil {
		return err
	}

	c.Put("state", bytes)
	return nil
}

// SetRGB sets the Nanoleaf's color by converting RGB to HSL.
func (c Client) SetRGB(red int, green int, blue int) error {
	h, s, l := rgbToHSL(red, green, blue)
	return c.SetHSL(h, s, l)
}

// BrightnessProperty represents the brightness of the Nanoleaf.
type BrightnessProperty struct {
	Value    int `json:"value"`
	Duration int `json:"duration,omitempty"`
}

// ColorTemperatureProperty represents the color temperature of the Nanoleaf.
type ColorTemperatureProperty struct {
	Value int `json:"value"`
}

// HueProperty represents the hue of the Nanoleaf.
type HueProperty struct {
	Value int `json:"value"`
}

// OnProperty represents the power state of the Nanoleaf.
type OnProperty struct {
	Value bool `json:"value"`
}

// SaturationProperty represents the saturation of the Nanoleaf.
type SaturationProperty struct {
	Value int `json:"value"`
}

// State represents a Nanoleaf state.
type State struct {
	On               *OnProperty               `json:"on,omitempty"`
	Brightness       *BrightnessProperty       `json:"brightness,omitempty"`
	ColorTemperature *ColorTemperatureProperty `json:"ct,omitempty"`
	Hue              *HueProperty              `json:"hue,omitempty"`
	Saturation       *SaturationProperty       `json:"sat,omitempty"`
}

// EffectsSelectRequest represents a JSON PUT body for `effects/select`.
type EffectsSelectRequest struct {
	Select string `json:"select"`
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
		Host:  cfg.Section("").Key("host").String(),
		Token: cfg.Section("").Key("access_token").String(),
	}

	if *verbose {
		fmt.Printf("Host: %s\n\n", client.Host)
	}

	if flag.NArg() > 0 {
		cmd := flag.Arg(0)
		switch cmd {
		case "off":
			state := State{
				On: &OnProperty{false},
			}
			bytes, err := json.Marshal(state)
			if err != nil {
				fmt.Printf("error: failed to marshal JSON: %v", err)
				os.Exit(1)
			}
			client.Put("state", bytes)
		case "on":
			state := State{
				On: &OnProperty{true},
			}
			bytes, err := json.Marshal(state)
			if err != nil {
				fmt.Printf("error: failed to marshal JSON: %v", err)
				os.Exit(1)
			}
			client.Put("state", bytes)
		case "hsl":
			doHSLCommand(client, flag.Args()[1:])
		case "rgb":
			doRGBCommand(client, flag.Args()[1:])
		case "effect":
			doEffectCommand(client, flag.Args()[1:])
		}
	}
}

func doEffectCommand(client Client, args []string) {
	if len(args) < 1 {
		fmt.Println("usage: picoleaf effect list")
		fmt.Println("       picoleaf effect select <name>")
		os.Exit(1)
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

func rgbToHSL(red, green, blue int) (int, int, int) {
	r := float64(red) / 255.0
	g := float64(green) / 255.0
	b := float64(blue) / 255.0

	min := math.Min(math.Min(r, g), b)
	max := math.Max(math.Max(r, g), b)

	c := max - min
	l := (max + min) / 2

	if c == 0 { // achromatic
		return 0, 0, int(math.Round(100 * l))
	}

	v := max

	h := 0.0
	switch v {
	case r:
		h = 0 + (g-b)/c
	case g:
		h = 2 + (b-r)/c
	case b:
		h = 4 + (r-g)/c
	}
	h *= 60
	if h < 0 {
		h += 360
	}

	s := (v - l) / math.Min(l, 1-l)

	return int(math.Round(h)), int(math.Round(100 * s)), int(math.Round(100 * l))
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
