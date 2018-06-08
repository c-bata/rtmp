# rtmp

Server implementation of RTMP 1.0 protocol in Go.

## Getting Started

Usage is like this:

```go
package main

import (
	"flag"
	"log"

	"github.com/c-bata/rtmp"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", ":1935", `TCP address to listen on, ":1935" if empty`)
	flag.Parse()

	log.Printf("Serving RTMP on %s", addr)
	err := rtmp.ListenAndServe(addr)
	if err != nil {
		log.Fatalf("Catch Error: %s", err)
	}
}
```

Build [an example server script](./_example/server/main.go) via Make and Run it.

```
$ make build
$ ./bin/server -addr :1935
2018/01/28 17:09:53 Serving RTMP on :1935 (rev-a669378)
```

After that, you can send a RTMP stream using RTMP client like ffmpeg or Wirecast.
In the case of using ffmpeg, please execute a following command.

```console
$ ffmpeg -re -i /path/to/your_video.mp4 -map 0 -c:v libx264 -c:a aac -f flv rtmp://127.0.0.1:1935/appName/appInstance
```

![GIF Animation - Receiving RTMP Stream.](https://github.com/c-bata/assets/raw/master/rtmp/rtmp-receiving-data-original.gif)

## Bibliography

* [RTMP 1.0 Specification - Adobe Systems Inc](http://www.adobe.com/devnet/rtmp.html)
* [Action Message Format 0 (AMF 0) specification - Adobe Systems Inc](http://wwwimages.adobe.com/content/dam/acom/en/devnet/pdf/amf0-file-format-specification.pdf)
* [Action Message Format 3 (AMF 3) specification - Adobe Systems Inc](http://wwwimages.adobe.com/content/dam/acom/en/devnet/pdf/amf-file-format-spec.pdf)
* [RTMP: A Quick Deep-Dive, Nick Chadwick - Youtube](https://www.youtube.com/watch?v=AoRepm5ks80)

