package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/russross/blackfriday"
)

const (
	finalTmpl = `
<html>
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=1024" />
	<meta name="apple-mobile-web-app-capable" content="yes" />
	<title>Show</title>

	<link href="{{with .CSS}}{{.}}{{else}}impress.css{{end}}" rel="stylesheet" />
</head>
<body class="impress-not-supported">
<div class="fallback-message">
	<p>Your browser <b>doesn't support the features required</b> by impress.js, so you are presented with a simplified version of this presentation.</p>
	<p>For the best experience please use the latest <b>Chrome</b>, <b>Safari</b> or <b>Firefox</b> browser.</p>
</div>
<div id="impress" data-transition-duration="250">
{{ .Content }}
</div>
<div class="hint">
	<p>Use a spacebar or arrow keys to navigate</p>
</div>
<script>
if ("ontouchstart" in document.documentElement) { 
	document.querySelector(".hint").innerHTML = "<p>Tap on the left or right to navigate</p>";
}
</script>
<script src="{{with .JS}}{{.}}{{else}}impress.js{{end}}"></script>
<script>impress().init();</script>
</body>
</html>
`
)

type pages struct {
	buf     *bytes.Buffer
	in, out chan string
}

var (
	filename = flag.String("f", "", "markdown file to process")
	css      = flag.String("css", "", "CSS filename")
	js       = flag.String("js", "", "impress.js filename")
)

func main() {
	flag.Parse()
	var r io.Reader
	if *filename == "" {
		r = os.Stdin
	} else {
		var err error
		r, err = os.Open(*filename)
		if err != nil {
			log.Fatal(err)
		}
	}
	scanner := bufio.NewScanner(r)
	p := &pages{
		buf: new(bytes.Buffer),
		out: make(chan string),
		in:  make(chan string),
	}
	go p.process()

	done := make(chan int)
	go output(p.out, done)

	for scanner.Scan() {
		line := scanner.Text()
		switch strings.TrimSpace(line) {
		case "---":
			p.emit()
		default:
			p.cache(line)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	if p.buf.String() != "" {
		p.emit()
	}
	p.end()

	<-done
}

func output(ch chan string, done chan int) {
	pageTmpl := `
<div class="slide step" data-x="{{.Xpixels}}">
{{.PageHTML}}
</div>
`
	tmpl := template.Must(template.New("").Parse(pageTmpl))
	buf := new(bytes.Buffer)
	n := 0
	for pageHTML := range ch {
		err := tmpl.Execute(buf, struct {
			Xpixels  int
			PageHTML string
		}{
			Xpixels:  n * 1000,
			PageHTML: pageHTML,
		})
		if err != nil {
			log.Fatal(err)
		}
		n++
	}
	tmpl = template.Must(template.New("").Parse(finalTmpl))
	err := tmpl.Execute(os.Stdout, struct {
		CSS, JS, Content string
	}{
		CSS:     *css,
		JS:      *js,
		Content: buf.String(),
	})
	if err != nil {
		log.Fatal(err)
	}
	done <- 1
}

func (p *pages) cache(line string) {
	_, err := p.buf.WriteString(line + "\n")
	if err != nil {
		log.Fatal(err)
	}
}

func (p *pages) emit() {
	p.in <- p.buf.String()
	p.buf.Reset()
}

func (p *pages) end() {
	close(p.in)
}

func (p *pages) process() {
	for page := range p.in {
		p.out <- transfer(page)
	}
	close(p.out)
}

func transfer(md string) string {
	return string(blackfriday.MarkdownCommon([]byte(md)))
}
