package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
)

// Client is a Nanoleaf REST API client.
type Client struct {
	Host  string
	Token string

	Verbose bool

	client http.Client
}

// Get performs a GET request.
func (c Client) Get(path string) (string, error) {
	if c.Verbose {
		fmt.Println("GET", path)
	}

	url := c.Endpoint(path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if c.Verbose {
		fmt.Println("<===", string(body))
		fmt.Println()
	}
	return string(body), nil
}

// Put performs a PUT request.
func (c Client) Put(path string, body []byte) error {
	if c.Verbose {
		fmt.Println("PUT", path)
		fmt.Println("===>", string(body))
		fmt.Println()
	}

	url := c.Endpoint(path)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	return nil
}

// Endpoint returns the full URL for an API endpoint.
func (c Client) Endpoint(path string) string {
	return fmt.Sprintf("http://%s/api/v1/%s/%s", c.Host, c.Token, path)
}

// ListEffects returns an array of effect names.
func (c Client) ListEffects() ([]string, error) {
	body, err := c.Get("effects/effectsList")
	if err != nil {
		return nil, err
	}

	var list []string
	err = json.Unmarshal([]byte(body), &list)
	return list, err
}

// Off turns off Nanoleaf.
func (c Client) Off() error {
	state := State{
		On: &OnProperty{false},
	}
	bytes, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return c.Put("state", bytes)
}

// On turns on Nanoleaf.
func (c Client) On() error {
	state := State{
		On: &OnProperty{true},
	}
	bytes, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return c.Put("state", bytes)
}

// SelectEffect activates the specified effect.
func (c Client) SelectEffect(name string) error {
	req := effectsSelectRequest{
		Select: name,
	}
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	c.Put("effects/select", bytes)
	return nil
}

// SetBrightness sets the Nanoleaf's brightness.
func (c Client) SetBrightness(brightness int) error {
	state := State{
		Brightness: &BrightnessProperty{brightness, 0},
	}

	bytes, err := json.Marshal(state)
	if err != nil {
		return err
	}

	c.Put("state", bytes)
	return nil
}

// SetColorTemperature sets the Nanoleaf's color temperature.
func (c Client) SetColorTemperature(temperature int) error {
	state := State{
		ColorTemperature: &ColorTemperatureProperty{temperature},
	}

	bytes, err := json.Marshal(state)
	if err != nil {
		return err
	}

	c.Put("state", bytes)
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

// effectsSelectRequest represents a JSON PUT body for `effects/select`.
type effectsSelectRequest struct {
	Select string `json:"select"`
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
