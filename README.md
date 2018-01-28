# rtmp

Server implementation of RTMP 1.0 protocol in Go.

## Usage

```go
package main

import (
	"log"
	"os"

	"github.com/c-bata/rtmp"
)

func main() {
	log.Print("Serving RTMP on :1935")

	err := rtmp.ListenAndServe(":1935")
	if err != nil {
		log.Printf("Got Error: %s", err)
		os.Exit(1)
	}
}
```

## Bibliography

* [RTMP 1.0 Specification - Adobe Systems Inc](http://www.adobe.com/devnet/rtmp.html)
* [Action Message Format 0 (AMF 0) specification - Adobe Systems Inc](http://wwwimages.adobe.com/content/dam/acom/en/devnet/pdf/amf0-file-format-specification.pdf)
* [Action Message Format 3 (AMF 3) specification - Adobe Systems Inc](http://wwwimages.adobe.com/content/dam/acom/en/devnet/pdf/amf-file-format-spec.pdf)
* [RTMP: A Quick Deep-Dive, Nick Chadwick - Youtube](https://www.youtube.com/watch?v=AoRepm5ks80)
