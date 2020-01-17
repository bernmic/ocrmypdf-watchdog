package main

import (
	"flag"
	"io/ioutil"
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
	log.Println("Watchdog started with:")
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
	// first get the parts of the path: dir+filename+ext
	directory := filepath.Dir(path)
	filename := filepath.Base(path)
	extension := filepath.Ext(filename)
	filename = filename[0 : len(filename)-len(extension)]
	// try to rename file
	tmpFile, err := ioutil.TempFile(directory, filename+".*."+extension)
	if err != nil {
		log.Printf("Unable to create temp file: %v", err)
		return
	}
	tmpFile.Close()
	os.Remove(tmpFile.Name())
	err = os.Rename(path, tmpFile.Name())
	if err != nil {
		log.Printf("Cannot rename file. Stopping here: %v", err)
		return
	}
	defer os.Remove(tmpFile.Name())

	target := c.OutFolder
	if !strings.HasSuffix(target, "/") {
		target = target + "/"
	}
	targetWithoutExtension := target + filename
	target = targetWithoutExtension + ".tmp"
	log.Printf("Run command >%s %s %s %s<\n", c.OCRMyPDFBinary, c.Parameter, tmpFile.Name(), target)
	runargs := strings.Split(c.Parameter, " ")
	runargs = append(runargs, tmpFile.Name(), target)
	cmd := exec.Command(c.OCRMyPDFBinary, runargs...)

	out, err := cmd.CombinedOutput()

	log.Println(string(out))

	log.Printf("Job finished with result %v\n", err)
	if err != nil {
		// error: tmp back to original name
		os.Rename(tmpFile.Name(), path)
	} else {
		// ok: rename tmp target to final target
		for fileExists(targetWithoutExtension+".pdf") {
			targetWithoutExtension+="_1"
		}
		os.Rename(target, targetWithoutExtension+".pdf")

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

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}
