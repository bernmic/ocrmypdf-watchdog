package main

import (
	"bufio"
	"flag"
	"fmt"
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
	PreScript      string
	PostScript     string
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
		os.Getenv("WATCHDOG_PRESCRIPT"),
		os.Getenv("WATCHDOG_POSTSCRIPT"),
	}
	flag.StringVar(&context.InFolder, "in", context.InFolder, "input folder")
	flag.StringVar(&context.OutFolder, "out", context.OutFolder, "output folder")
	flag.StringVar(&context.OCRMyPDFBinary, "ocrmypdf", context.OCRMyPDFBinary, "ocrmydpf binary to use")
	flag.IntVar(&context.Frequency, "frequency", frequency, "frequency in seconds")
	flag.StringVar(&context.PreScript, "prescript", context.PreScript, "name of the script to run before each document")
	flag.StringVar(&context.PostScript, "postscript", context.PostScript, "name of the script to run after each document")
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
	log.Println("PreScript = " + context.PreScript)
	log.Println("PostScript = " + context.PostScript)

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
		cc := c.readPropertiesFiles()
		for _, file := range files {
			cc.processDocument(file)
		}

		timer := time.NewTimer(frequency)
		<-timer.C
		timer.Stop()
	}
}

func (c *Context) processDocument(path string) {
	runScript(c.InFolder, c.PreScript)
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
		for fileExists(targetWithoutExtension + ".pdf") {
			targetWithoutExtension += "_1"
		}
		os.Rename(target, targetWithoutExtension+".pdf")

	}
	runScript(c.OutFolder, c.PostScript)
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

func runScript(path string, script string) {
	if script != "" {
		s := fmt.Sprintf("%s/%s", path, script)
		if !fileExists(s) {
			log.Printf("WARN: could not find script %s\n", s)
			return
		}
		log.Printf("run script %s\n", s)
		cmd := exec.Command("bash", "-c", s)
		out, err := cmd.CombinedOutput()
		log.Println(string(out))
		log.Printf("Job finished with result %v\n", err)
	}
}

func (c *Context) readPropertiesFiles() *Context {
	var files []string
	err := filepath.Walk(c.InFolder, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if strings.HasSuffix(path, ".properties") {
				files = append(files, path)
			}
		}
		return nil
	})
	if err == nil && len(files) > 0 {
		m, err := readPropertiesFile(files[0])
		if err != nil {
			return c
		}
		cc := *c
		for k, v := range m {
			if k == "OCRMYPDF_PARAMETER" {
				cc.Parameter = v
			} else if k == "WATCHDOG_EXTENSIONS" {
				cc.Extensions = v
			}
		}
		return &cc
	}
	return c
}

func readPropertiesFile(filename string) (map[string]string, error) {
	config := make(map[string]string, 0)

	if len(filename) == 0 {
		return config, nil
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				config[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		return nil, err
	}

	return config, nil
}
