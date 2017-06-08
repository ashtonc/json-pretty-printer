This is a simple JSON pretty printer written in Go for a school assignment. It does not use the [encoding/json](https://golang.org/pkg/encoding/json/) package. The assignment was intended to teach use the basics of parsing and lexical analysis.

To use it, simply run it on the command line with a JSON input as the first argument. By default, the HTML output is sent to stdout, so if you want to save it you should redirect it to an HTML file (eg. go run json-pretty-printer.go input.json > output.html).

The HTML output decorates the JSON with a hard-coded color scheme. The program also fails if the input is not valid JSON. I unfortunately lost the original git repository that with my development history for the project, so for now it is simply one commit set to the project submission time.

