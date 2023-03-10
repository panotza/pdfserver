package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"pdfserver/chrome"
	"pdfserver/pdf"
)

func main() {
	headlessChrome, err := chrome.NewChrome(1000)
	if err != nil {
		log.Panic(err)
	}
	defer headlessChrome.Close()

	pdfRenderer := pdf.NewChrome(headlessChrome)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		buf, err := pdfRenderer.SaveAsPDF(r.Context(), `<!DOCTYPE html>
		<html>
		<head>
		<title>Page Title</title>
		</head>
		<body>
		
		<h1>This is a Heading</h1>
		<p>This is a paragraph.</p>
		
		</body>
		</html>`)
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(w, bytes.NewReader(buf))
		if err != nil {
			panic(err)
		}
	})

	http.ListenAndServe(":3000", nil)
}
