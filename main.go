package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Context struct {
	InFolder       string
	OutFolder      string
	OCRMyPDFBinary string
	Parameter      string
	Frequency      int
	Extensions     string
}

func main() {
	frequency, err := strconv.Atoi(os.Getenv("WATCHDOG_FREQUENCY"))
	if err != nil {
		frequency = 1
	}
	context := &Context{
		os.Getenv("OCRMYPDF_IN"),
		os.Getenv("OCRMYPDF_OUT"),
		os.Getenv("OCRMYPDF_BINARY"),
		os.Getenv("OCRMYPDF_PARAMETER"),
		frequency,
		os.Getenv("WATCHDOG_EXTENSIONS"),
	}
	flag.StringVar(&context.InFolder, "in", context.InFolder, "input folder")
	flag.StringVar(&context.OutFolder, "out", context.OutFolder, "output folder")
	flag.StringVar(&context.OCRMyPDFBinary, "ocrmypdf", context.OCRMyPDFBinary, "ocrmydpf binary to use")
	flag.IntVar(&context.Frequency, "frequency", frequency, "frequency in seconds")

	flag.Parse()

	if context.InFolder == "" || context.OutFolder == "" {
		log.Fatalln("in and/or out folder not defined.")
	}
	if context.OCRMyPDFBinary == "" {
		context.OCRMyPDFBinary = "ocrmypdf"
	}
	if context.Parameter == "" {
		context.Parameter = "-l eng+fra+deu --rotate-pages --deskew --jobs 4 --output-type pdfa"
	}
	if context.Extensions == "" {
		context.Extensions = "pdf,tif,tiff,jpg,jpeg,png,gif"
	}
	log.Println("Watchdog startet with:")
	log.Println("in = " + context.InFolder)
	log.Println("out = " + context.OutFolder)
	log.Printf("Frequency = %d seconds\n", context.Frequency)
	log.Println("Extensions to look for: " + context.Extensions)
	log.Println("OCRMyPDF binary = " + context.OCRMyPDFBinary)
	log.Println("OCRMyPDF parameter = " + context.Parameter)

	context.watchdog()
}

func (c *Context) watchdog() {
	frequency := time.Duration(c.Frequency) * time.Second
	for {
		var files []string
		err := filepath.Walk(c.InFolder, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				if c.hasOneOfExtensions(path) {
					files = append(files, path)
				}
			}
			return nil
		})
		if err != nil {
			panic(err)
		}
		for _, file := range files {
			c.processDocument(file)
		}

		timer := time.NewTimer(frequency)
		<-timer.C
		timer.Stop()
	}
}

func (c *Context) processDocument(path string) {
	log.Println("Processing file " + path)
	// check if path is not open by another process
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_EXCL, 0)
	if err != nil {
		log.Printf("File %s not ready. Stop here.", path)
		return
	}
	f.Close()
	filename := filepath.Base(path)
	target := c.OutFolder
	if !strings.HasSuffix(target, "/") {
		target = target + "/"
	}
	target = target + filename
	runArgs := fmt.Sprintf(`%s '%s' '%s'`, c.Parameter, path, target)
	log.Printf("Run command >%s %s<\n", c.OCRMyPDFBinary, runArgs)
	cmd := exec.Command(c.OCRMyPDFBinary, runArgs)

	var out bytes.Buffer
	cmd.Stdout = &out
	var ser bytes.Buffer
	cmd.Stderr = &ser

	err = cmd.Run()

	log.Println(out.String())
	log.Println(ser.String())

	log.Printf("Job finished with result %v\n", err)
	if err := os.Remove(path); err != nil {
		log.Printf("Error removing path %s: %v", path, err)
	}
}

func (c *Context) hasOneOfExtensions(path string) bool {
	extensions := strings.Split(c.Extensions, ",")
	for _, s := range extensions {
		if strings.HasSuffix(strings.ToLower(path), "."+s) {
			return true
		}
	}
	return false
}
