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
* WATCHDOG_PRESCRIPT
* WATCHDOG_POSTSCRIPT

## Parameters

The watchdog accepts the following parameters:

* --in <in-path>
* --out <out-path>
* --frequency <in seconds)
* --ocrmypdf <path and name of the executable>
* --prescript <name of script>
* --postscript <name of script>

## Pre- and postprocessing

The watchdog allows pre- and postprocessing for each document.

A preprocessing script can be copied to ```in``` folder and the name of the script must be added by the --prescript parameter or the WATCHDOG_PRESCRIPT environment variable. It will be executed for each document just before ocrmypdf is called.

A postprocessing script can be copied to ```out``` folder and the name of the script must be added by the --postscript parameter or the WATCHDOG_POSTSCRIPT environment variable. It will be executed for each document after ocrmypdf is called.

## Temporary parameters

The watchdog allows to overwrite temporarely the commandline parameters. If a file with extension ```.properties``` is found in the ```in``` folder which contains keys and values separated through an equal sign, it will overwrite the value.

### Example
```OCRMYPDF_PARAMETER=-l eng+fra+deu --rotate-pages --deskew --jobs 4 --output-type pdfa```

For now only WATCHDOG_EXTENSIONS and OCRMYPDF_PARAMETER are considered.

## Multi architecture build

```docker buildx build -t "${DOCKER_USER}/ocrmypdf-watchdog:latest" --platform linux/amd64,linux/arm64 --push .```
