[![Build Status](https://travis-ci.org/ausocean/h264decode.svg?branch=master)](https://travis-ci.org/ausocean/h264decode) [![Go Report Card](https://goreportcard.com/badge/github.com/ausocean/h264decode)](https://goreportcard.com/report/github.com/ausocean/h264decode)

Listens for a stream of H264 bytes on port 8000.

A stream is read sequentially dropping each NAL into a struct with access to the RBSP and seekable features. No interface contracts are implemented right now. This is heavily a work in progress.

# TODO

* CABAC initialization
* Context-adaptive arithmetic entropy-coded syntax element support
* Macroblock to YCbCr image decoding

## Done

* DecodeBypass, 9.3.3.2.3
* DecodeTerminate, 9.3.3.2.4
* RenormD - 9.3.3.2.2
* rangeTableLPS ( Table 9-44 )
* Derive ctxIDX per 9.3.3.1
* Select M, N values

### ArithmeticDecoding S 9.3.3.2

* cabac.go : ArithmeticDecoding 9.3.3.3.2
* cabac.go : DecodeBypass, DecodeTerminate, DecodeDecision

## In Progress

* Make use of DecodeBypass and DecodeTerminate information
* 9.3.3.2.1 - BinaryDecision
 * 9.3.3.2.1.1, 9.3.3.2.2

## Next

* Make use of initCabac (initialized CABACs)

# Background

The last point was and is the entire driving force behind this project: To decode a single frame to an image and begin doing computer vision tasks on it. A while back, this project was started to keep an eye on rodents moving their way around various parts of our house and property. What was supposed to happen was motion detected from one-frame to another of an MJPEG stream would trigger capturing the stream. Analyzing the stream, even down at 24 fps, caused captures to be triggered too late. When it was triggered, there was so much blur in the resulting captured stream, it wasn't very watchable.

Doing a little prototyping it was apparent reading an h.264 stream into VLC provided a watchable, unblurry, video. With a little searching on the internet, (https://github.com/gqf2008/codec) provided an example of using [ffMPEG](https://www.ffmpeg.org/) to decode an h.264 frame/NAL unit's macroblocks into a YCbCr image. Were this some shippable product with deadlines and things, [GGF2008](https://github.com/gqf2008/codec) would've been wired up and progress would've happened. But this is a tinkering project. Improving development skills, learning a bit about streaming deata, and filling dead-air were the criteria for this project.

Because of that, a pure Go h.264 stream decoder was the only option. Duh.

