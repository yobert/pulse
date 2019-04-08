# NOTE:
Someone nice has rewritten this package into a much better package, with better code, a better API, and more protocol coverage! See https://github.com/jfreymuth/pulse

# pulse
PulseAudio client implementation in pure Go

# why?
I saw several wrappers out there of libpulse in Go using CGO, but CGO has many drawbacks. I wanted to try actually implementing the pulseaudio wire protocol in Go instead. The protocol is a very nice binary protocol that can speak over a unix socket. Mostly I figured it out by reading pulseaudio's source code, and by a little debugging unix socket proxy that I've included here. Just fire up the proxy and then run PULSE_SERVER="unix:/tmp/pulsedebug" paplay something.wav

# status

Working:
- Basic protocol negotiation
- It will connect to the default audio sink and synthesize a 440hz (A) sine wave

Not working yet:
- It does not correctly buffer according to the servers requests (It just sends frames on a time schedule and hopes it comes out unbroken).
- shm / memfd support
- The rest of the protocol...
