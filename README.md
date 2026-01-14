# PixelBox

PixelBox is software to control the Divoom Timebox Evo. You can dynamically show
images and information on the device as you see fit. You can use the server to
control most aspects of the Timebox over HTTP.

For example, you could write a little script yourself or create Home Assistant
automations, or maybe plug it in to your favourite streaming software. Show the
current weather, your YouTube subscriber count, the price of Bitcoin or when
someone's at the door. The sky is now the limit.

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
figured out the commands I needed from the resources above and messed with the
implementation until it worked. I think the Go code in this repository is simple
enough for it to be a reference for people wanting to use my knowledge.

There's just two parts to it; the message sent and the wrapping of said message.
The wrapping is in [`envelope.go`](./divoom/envelope.go). The messages to send
to the speaker are constructed by [`outgoing.go`](./divoom/outgoing.go). I've
made fairly little effort to decode the messages coming back from the device,
but what I figured out lives in [`incoming.go`](./divoom/incoming.go).
