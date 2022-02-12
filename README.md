# ocrmypdf-watchdog

This is a simple watchdog for OCRMyPDF (and maybe others). It watches a given folder for new files with definable extensions and runs then ocrmypdf (or another command) to convert files to pdf.

## Docker

The Dockerfile creates an image based on the jbarlow83/ocrmypdf image and adds the watchdog.

The docker-compose creates a container from the image. The first time it has to be started with the --build flag to build the image:

    docker-compose up --build
 
 There are 2 volumes: <b>/in</b> and <b>/out</b>
 The docker-compose.yml shows how to use them.
 
 ## Environment
 
The watchdog looks for the following environment variables:
 
* OCRMYPDF_IN
* OCRMYPDF_OUT
* OCRMYPDF_BINARY
* OCRMYPDF_PARAMETER
* WATCHDOG_EXTENSIONS
* WATCHDOG_FREQUENCY

## Parameters

The watchdog accepts the following parameters:

* --in <in-path>
* --out <out-path>
* --frequency <in seconds)
* --ocrmypdf <path and name of the executable>

## Multi architecture build

docker buildx build -t "${DOCKER_USER}/ocrmypdf-watchdog:latest" --platform linux/amd64,linux/arm64 --push .