# impress-md

Impress-md is a tool to convert markdown syntax file into [impress.js](https://github.com/bartaz/impress.js/) presentation html file. It recognizes the `---` separator in your markdown file, sends segments to [blackfriday](https://github.com/russross/blackfriday), then generates a single html file and writes it to your standard output. You can specify the location of `impress.js` and custom css file.

## Install

After [installing Go](http://golang.org/doc/install) and [setting your Go environment](http://golang.org/doc/code.html), just `go get` it:

	go get github.com/fanyang01/impress-md
