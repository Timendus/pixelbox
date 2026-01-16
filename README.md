# PixelBox

PixelBox is software to control the Divoom Timebox Evo over the network. You can
create scenes using the web-based user interface. Scenes can show images,
animations, visualisations and information on the device as you see fit. You can
then manually switch scenes, or by making HTTP calls to the server.

For example, you could create Home Assistant automations to switch scenes, or
maybe plug it in to your favourite streaming software. You can also POST images
or animated GIF files to the server to show those on the device. With some
creativity and a bit of scripting you can show the current weather, your YouTube
subscriber count, the price of Bitcoin or when someone's at the door. That's all
up to your imagination.

## How to use

Besides a Divoom Timebox Evo, you need a device to run this software on that has
both networking and Bluetooth. It was designed with a Raspberry Pi in mind, but
any computer should do. This computer connects to the speaker over Bluetooth and
runs the server to control it over the network.

## Timebox Evo Bluetooth Protocol

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
