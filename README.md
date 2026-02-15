# PixelBox

## The backstory

I recently had the flu and was pretty sick for two weeks. I needed a little
project to keep my mind off of the headache I couldn't shake. My eyes fell on
the Divoom Timebox Evo that was still sitting on my shelf, and started messing
with it.

When I won it in a little internal dev jam that we organised at my employer's, I
had told myself I would free it from its dependency on its proprietary app. That
was years ago of course; something more interesting or more pressing always got
priority. But this was about the right amount of complexity and
straightforwardness that I thought I needed to pull myself through these two
weeks, so I started hacking.

## The software

PixelBox is software to control the Divoom Timebox Evo over the network. You can
create scenes using the web-based user interface. Scenes can show images,
animations, visualisations and information on the device as you see fit. You can
then manually switch scenes, or by making HTTP calls to the server.

For example, you could create Home Assistant automations to switch scenes, or
maybe plug it in to your favourite streaming software. You can also POST images
or animated GIF files to the server to show those on the device. With some
creativity and a bit of scripting you can show the current weather, your YouTube
subscriber count, the price of Bitcoin or when someone's at the door. That's all
up to your imagination. This software only exposes the device.

## How to use

Besides a Divoom Timebox Evo, you need a device to run this software, one that
has both networking and Bluetooth. It was designed with a Raspberry Pi in mind,
but any Linux machine should do. This device connects to the speaker over
Bluetooth and runs the server to control it over the network.

### Downloading Pixelbox

You can download the latest release from the [releases
page](https://github.com/Timendus/pixelbox/releases) on Github. Or on a
Raspberry Pi you can do something like this:

```bash
mkdir pixelbox
cd pixelbox
wget https://github.com/Timendus/pixelbox/releases/latest/download/pixelbox-linux-arm.tar.gz
tar -xvf pixelbox-linux-arm.tar.gz
```

### Connecting to the speaker

#### On a Linux Desktop

Just use the graphical user interface to connect to the speaker. If the UI
doesn't give you the MAC address of the speaker, you can use this command to
find it:

```bash
bluetoothctl devices
```

#### On a Raspberry Pi

If you're working with a headless Raspberry Pi, like I chose to do,
unfortunately I think you will find that it is much more cumbersome to get the
pi to connect to the Timebox Evo over Bluetooth than it is to run this software.
I don't have a perfect manual for you, but it's something like this:

```bash
sudo apt update
sudo apt install pipewire pipewire-pulse wireplumber libspa-0.2-bluetooth pipewire-audio
sudo systemctl enable bluetooth
sudo systemctl start bluetooth
sudo rfkill unblock bluetooth
bluetoothctl
```

Then in `bluetoothctl`:

```bluetoothctl
power on
agent on
default-agent
scan on
pair XX:XX:XX:XX:XX:XX
trust XX:XX:XX:XX:XX:XX
connect XX:XX:XX:XX:XX:XX
```

And if you also want to use the speaker for audio:

```bash
loginctl enable-linger $USER
sudo mkdir -p /etc/wireplumber/wireplumber.conf.d
sudo tee /etc/wireplumber/wireplumber.conf.d/51-bluez-seat.conf >/dev/null <<'EOF'
wireplumber.profiles = {
  main = {
    monitor.bluez.seat-monitoring = disabled
  }
}
EOF
systemctl --user restart wireplumber
systemctl --user restart pipewire
```

This is a very rough sketch. Properly describing this is outside the scope of
this software. Use Google and your favourite AI to work your way through this :)

### Running PixelBox

Configure PixelBox by editing `config.json` with your favourite editor. Put the
Bluetooth MAC address of the speaker that you discovered in the previous step in
the `mac` field of the device. If you want, change the port on which the web
service will run. The rest should be fine.

You can then either just run the `pixelbox` binary from its directory or install
PixelBox as a systemd service, so it runs in the background and starts at boot.
This is how you do the latter:

```bash
nano pixelbox.service
# Fix the path to where you downloaded pixelbox in both places, then run:
./install.sh
```

Once the service is running, open the web UI in a web browser to control the
Timebox and configure your scenes. It should be running on `http://<ip or
hostname of your pi>:3000`.

### Image support

If you want to be able to upload images and animated GIF files to PixelBox, it
needs to be served over HTTPS. The quickest way to do this on a Raspberry Pi is
to install Caddy as a reverse proxy server.

You can use the snippet below to do that, but make sure you change the IP
address to the IP of your pi and if you changed the port in `config.json`,
you'll have to change the port on the line that starts with `reverse_proxy` to
match.

```bash
sudo apt install -y caddy
sudo cp /etc/caddy/Caddyfile /etc/caddy/Caddyfile-backup
sudo tee /etc/caddy/Caddyfile >/dev/null <<'EOF'
https://192.168.1.105 {
  tls internal
  reverse_proxy 127.0.0.1:3000
}
EOF
sudo systemctl restart caddy
```

You should now be able to access PixelBox on `https://<ip or hostname of your
pi>` directly.

### Using the API

For scripting, you can use the following endpoints:

- `GET /scene/<id>/apply` - Apply the scene with the given ID
- `GET /apply/syncTime` - Send the system time to the Timebox
- `POST /apply/image` - Show the given static image
- `POST /apply/gif` - Show the given animated GIF file

The two image endpoints expect a multipart form with a field called `file`,
which holds the image. Here's an example of a snippet of HTML that you can use
to show an image:

```html
<form action="/apply/image" method="POST" enctype="multipart/form-data">
  <input type="file" name="file" accept="image/png, image/jpeg, image/gif" />
  <button type="submit">Show image</button>
</form>
```

## Timebox Evo Bluetooth Protocol

I didn't have to reverse engineer everything myself, which made this project
feasible to do in a couple of weeks.

I've mostly used the resource below. Although the author clearly has a limited
grasp of working with binary data, instead treating everything like strings, it
was the most correct when it came to the protocol.

https://github.com/RomRider/node-divoom-timebox-evo/blob/master/PROTOCOL.md#visualization

I've also consulted this, which deals with a slightly different device, and is
also a bit cryptic at times:

https://github.com/MarcG046/timebox/blob/master/doc/protocol.md

There are more resources linked by both authors.

I will not try yet another attempt to document the protocol myself. I've only
figured out the commands from the resources above and messed with the
implementation until it worked. I think the Go code in this repository is simple
enough for it to be a reference for people wanting to use my knowledge.

There's just two parts to it; the message sent and the wrapping of said message.
The wrapping is in [`envelope.go`](./protocol/envelope.go). The messages to send
to the speaker are constructed by [`outgoing.go`](./protocol/outgoing.go). I've
not made an aweful lot of effort to decode the messages coming back from the
device, but what I figured out lives in [`incoming.go`](./protocol/incoming.go).

## Developing

In short: Make sure you have [Go](https://go.dev/doc/install) installed, check
out the project and run `make`.

```bash
git clone git@github.com:Timendus/pixelbox.git
cd pixelbox
go mod install  # Not sure you even need this
make
```
