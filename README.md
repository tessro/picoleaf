# Picoleaf

Picoleaf is a tiny CLI tool for controlling Nanoleaf.

## Installation

Make sure `$GOPATH/bin` is on your `$PATH`, then run:

```bash
go install github.com/paulrosania/picoleaf
```

Picoleaf expects a `.picoleafrc` file in your home directory, with the
following settings:

```ini
host=<ip address>:<port>
access_token=<token>
```

You can find your Nanoleaf's IP address via your router console. Your Nanoleaf's
port is probably `16021`.

To create an access token, you'll need to do the following:

1. On your Nanoleaf controller, hold the on-off button for 5-7 seconds until the
   LED starts flashing in a pattern.
2. Within 30 seconds, run: `curl -iLX POST http://<ip address>:<port>/api/v1/new`

This should print a token to your console.


## Usage

```bash
# Power
picoleaf on   # Turn Nanoleaf on
picoleaf off  # Turn Nanoleaf off

# Colors
picoleaf hsl <hue> <saturation> <lightness>  # Set Nanoleaf to the provided HSL
picoleaf rgb <red> <green> <blue>            # Set Nanoleaf to the provided RGB
picoleaf temp <temperature>                  # Set Nanoleaf to the provided color temperature
picoleaf brightness <temperature>            # Set Nanoleaf to the provided brightness

# Effects
picoleaf effect list           # List installed effects
picoleaf effect select <name>  # Activate the named effect
```
