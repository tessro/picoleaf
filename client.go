package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
)

// ExternalControlPort is the UDP port for Nanoleaf external control.
const ExternalControlPort = 60222

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
func (c Client) Put(path string, body []byte) (string, error) {
	if c.Verbose {
		fmt.Println("PUT", path)
		fmt.Println("===>", string(body))
	}

	url := c.Endpoint(path)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if c.Verbose {
		fmt.Println("<===", res.Status)
		if len(responseBody) > 0 {
			fmt.Println("<===", string(responseBody))
		}
		fmt.Println()
	}
	return string(responseBody), nil
}

// Endpoint returns the full URL for an API endpoint.
func (c Client) Endpoint(path string) string {
	return fmt.Sprintf("http://%s/api/v1/%s/%s", c.Host, c.Token, path)
}

// Effects represents the Nanoleaf panel effects state.
type Effects struct {
	Selected string   `json:"select"`
	List     []string `json:"effectsList"`
}

// Rhythm represents the Nanoleaf rhythm state.
type Rhythm struct {
	Connected       bool   `json:"rhythmConnected"`
	Active          bool   `json:"rhythmActive"`
	ID              int    `json:"rhythmId"`
	HardwareVersion string `json:"hardwareVersion"`
	FirmwareVersion string `json:"firmwareVersion"`
	AuxAvailable    bool   `json:"auxAvailable"`
	Mode            int    `json:"rhythmMode"`
	Position        struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		O float64 `json:"o"`
	} `json:"rhythmPos"`
}

// PanelLayout represents the Nanoleaf panel layout.
type PanelLayout struct {
	Layout struct {
		NumPanels    int `json:"numPanels"`
		SideLength   int `json:"sideLength"`
		PositionData []struct {
			PanelID   int `json:"panelId"`
			X         int `json:"x"`
			Y         int `json:"y"`
			O         int `json:"o"`
			ShapeType int `json:"shapeType"`
		} `json:"positionData"`
	} `json:"layout"`
	GlobalOrientation struct {
		Value int `json:"value"`
		Max   int `json:"max"`
		Min   int `json:"min"`
	} `json:"globalOrientation"`
}

// PanelInfo represents the Nanoleaf panel info response.
type PanelInfo struct {
	Name            string      `json:"name"`
	SerialNo        string      `json:"serialNo"`
	Manufacturer    string      `json:"manufacturer"`
	FirmwareVersion string      `json:"firmwareVersion"`
	Model           string      `json:"model"`
	State           State       `json:"state"`
	Effects         Effects     `json:"effects"`
	PanelLayout     PanelLayout `json:"panelLayout"`
	Rhythm          Rhythm      `json:"rhythm"`
}

// GetPanelInfo returns the Nanoleaf panel info.
func (c Client) GetPanelInfo() (*PanelInfo, string, error) {
	body, err := c.Get("")
	if err != nil {
		return nil, "", err
	}

	var panelInfo PanelInfo
	err = json.Unmarshal([]byte(body), &panelInfo)
	return &panelInfo, string(body), err
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
	_, err = c.Put("state", bytes)
	return err
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
	_, err = c.Put("state", bytes)
	return err
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
		Brightness: &BrightnessProperty{Value: brightness},
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
		ColorTemperature: &ColorTemperatureProperty{Value: temperature},
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
		Brightness: &BrightnessProperty{Value: lightness},
		Hue:        &HueProperty{Value: hue},
		Saturation: &SaturationProperty{Value: sat},
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

// startExternalControl sets Nanoleaf to accept UDP input.
func (c Client) startExternalControl() error {
	_, err := c.Put("effects", []byte(`{"write":{"command":"display","animType":"extControl","extControlVersion":"v2"}}`))
	return err
}

// SetPanelColor represents a frame of external color data.
type SetPanelColor struct {
	PanelID        uint16
	Red            uint8
	Green          uint8
	Blue           uint8
	White          uint8
	TransitionTime uint16
}

// SetCustomColors sets individual Nanoleaf pane colors.
func (c Client) SetCustomColors(frames []SetPanelColor) error {
	err := c.startExternalControl()
	if err != nil {
		return err
	}

	hostAddr, err := net.ResolveTCPAddr("tcp", c.Host)
	if err != nil {
		return err
	}

	laddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return err
	}

	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", hostAddr.IP, ExternalControlPort))
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		return err
	}

	numPanels := len(frames)
	if numPanels < 0 || numPanels > math.MaxUint16 {
		return fmt.Errorf("Expected between 0-%d panels, got %d", math.MaxUint16, numPanels)
	}

	headerSize := 2
	panelFrameSize := 8
	controlFrameSize := headerSize + panelFrameSize*numPanels
	buf := make([]byte, controlFrameSize)
	binary.BigEndian.PutUint16(buf, uint16(numPanels))
	for i, panel := range frames {
		offset := headerSize + panelFrameSize*i
		binary.BigEndian.PutUint16(buf[offset:], panel.PanelID)
		buf[offset+2] = panel.Red
		buf[offset+3] = panel.Green
		buf[offset+4] = panel.Blue
		buf[offset+5] = panel.White
		binary.BigEndian.PutUint16(buf[offset+6:], panel.TransitionTime)
	}

	conn.Write(buf)
	conn.Close()
	return nil
}

// BrightnessProperty represents the brightness of the Nanoleaf.
type BrightnessProperty struct {
	Min      *int `json:"min,omitempty"`
	Max      *int `json:"max,omitempty"`
	Value    int  `json:"value"`
	Duration int  `json:"duration,omitempty"`
}

// ColorTemperatureProperty represents the color temperature of the Nanoleaf.
type ColorTemperatureProperty struct {
	Min   *int `json:"min,omitempty"`
	Max   *int `json:"max,omitempty"`
	Value int  `json:"value"`
}

// HueProperty represents the hue of the Nanoleaf.
type HueProperty struct {
	Min   *int `json:"min,omitempty"`
	Max   *int `json:"max,omitempty"`
	Value int  `json:"value"`
}

// OnProperty represents the power state of the Nanoleaf.
type OnProperty struct {
	Value bool `json:"value"`
}

// SaturationProperty represents the saturation of the Nanoleaf.
type SaturationProperty struct {
	Min   *int `json:"min,omitempty"`
	Max   *int `json:"max,omitempty"`
	Value int  `json:"value"`
}

// State represents a Nanoleaf state.
type State struct {
	On               *OnProperty               `json:"on,omitempty"`
	Brightness       *BrightnessProperty       `json:"brightness,omitempty"`
	ColorTemperature *ColorTemperatureProperty `json:"ct,omitempty"`
	Hue              *HueProperty              `json:"hue,omitempty"`
	Saturation       *SaturationProperty       `json:"sat,omitempty"`
	ColorMode        string                    `json:"colorMode,omitempty"`
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
