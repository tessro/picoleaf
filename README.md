# Picoleaf

Picoleaf is a tiny CLI tool for controlling Nanoleaf.

## Installation

### macOS

`picoleaf` is available via a Homebrew Tap:

```bash
brew install paulrosania/command-home/picoleaf
```

You can also download a precompiled binary from the
[releases](https://github.com/paulrosania/picoleaf/releases) page.

### Source

Make sure Go is installed, and that `$GOPATH/bin` is on your `$PATH`. Then run:

```bash
go install github.com/paulrosania/picoleaf
```

# Getting Started

Picoleaf expects a `.picoleafrc` file in your home directory, with the
following settings:

```ini
host=<hostname or ip address>:<port>
access_token=<token>
```

## Manual creation of .picoleafrc

You can find your Nanoleaf's IP address via your router console. Your Nanoleaf's
port is probably `16021`.

Alternatively, you may be able to use mDNS service discovery. For example, on
macOS you can do the following:

```bash
$ dns-sd -Z _nanoleafapi | grep -o '\w*\-.*\.local'

# => 16021 Nanoleaf-Light-Panels-xx-xx-xx.local
#
# Use this as your `host` setting. Don't forget to append the port number.
#
# (You'll need to Ctrl-C to wrap up, since `dns-sd` listens indefinitely.)
```

To create an access token, you'll need to do the following:

1. On your Nanoleaf controller, hold the on-off button for 5-7 seconds until the
   LED starts flashing in a pattern.
2. Within 30 seconds, run: `curl -iLX POST http://<ip address>:<port>/api/v1/new`

This should print a token to your console.

Create and edit `~/.picoleafrc` with the values you have discovered.

## Use of the create picoleafrc helper 
This tiny script will put together your `.picoleafrc` file for you.

1. On your Nanoleaf controller, hold the on-off button for 5-7 seconds until the
   LED starts flashing in a pattern.
2. Within 30 seconds, run: `./create_picoleafrc > ~/.picoleafrc`

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
picoleaf effect custom [<panel> <red> <green> <blue> <transition time>] ...

# Panel properties
picoleaf panel info     # Print all panel information
picoleaf panel model    # Print Nanoleaf model
picoleaf panel name     # Print Nanoleaf name
picoleaf panel version  # Print Nanoleaf and rhythm module versions
```
