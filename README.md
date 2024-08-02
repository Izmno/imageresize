# ImageResize Thumbnail Microservice

## Overview

ImageResize is a simple thumbnail microservice designed to serve images from 
disk, resizing them to thumbnail size using ImageMagick before returning them. 
This microservice is packaged with a Dockerfile for easy deployment, and the 
Docker image is hosted on Dockerhub. Configuration options are available via 
environment variables, and a sample Docker Compose file is provided to help you 
get started quickly.

## Features

- Serve images from disk
- Resize images to thumbnail size using ImageMagick
- Configurable via environment variables
- Dockerized for easy deployment
- Example Docker Compose file included

## Prerequisites

- Docker
- Docker Compose (optional, for using the provided example)

## Installation

### Docker

Pull the docker image 

```
docker pull simondemeyere/imageresize
```

### Build from Source

To build the Docker image from source:

```
git clone git@github.com:izmno/imageresize.git

cd imageresize
docker build -t imageresize .
```

## Usage

### Running the Service

You can run the service using Docker:

```
docker run -d \
    -p 8080:8080 \
    -v ./examples/data:/media \
    simondemeyere/imageresize
```

### Configuration

ImageResize can be configured using environment variables:

Variable   | Description
---------- | -----------
PORT       | Port to listen on (default: 8080)
MEDIA_PATH | Path of media directory to serve images from (default: /media)

### Example Docker Compose

An example Docker Compose file is included to help you get started quickly.

```
docker compose -f ./examples/docker-compose.yaml up --build 
```

Two example images are included in `./examples/data`.

```
curl http://localhost:8080/dice.png
curl http://localhost:8080/balls.jpg
```


# Contributing

We welcome contributions! Please fork the repository and submit a pull request 
for any improvements or bug fixes.

# License

This project is licensed under the GPLv3 License. See the LICENSE file for details.

# Contact

For questions or support, please open an issue on the GitHub repository.
