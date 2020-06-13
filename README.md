# NEFAS [![Build Status](https://travis-ci.org/dff-labs/nefas.svg?branch=master)](https://travis-ci.org/mzahmi/ventilator)

## Overview

This is the code for the [Dubai Future Foundation Open Source Ventilator](https://m061.dubaifuture.gov.ae).

This project drives the main ventilator hardware, and is written in Go. 

A GUI is implemented in Python-QT, and can be imported as a submodule from https://gitlab.com/ralsuwaidi/new-ui.

The overall application architecture is as follows:
* Ventilator (hardware) is accessed via a number of sensors talking I2C and SPI, accessed via GPIO 
* The Go application reads data from the ventilator and stores readings to a Redis database
* Redis database is subsequently accessible by any other front-end (the GUI application)

Yes, we probably need a pretty diagram for this. That will come soon!

The project (including the submodules) use the following technologies:
* Golang    - the core application talking to the sensors is written in Go
* Redis     - to store data from the sensors
* Python    - for reading and processing data
* PySide2 (QT) - for the GUI

## Development 

The core application runs on Raspberry Pi Buster, so basically Debian. So if you want to contribute, it's probably best if you run things on a Debian/Ubuntu image. However, for testing purposes the application can be built on an other OS just as well. You just need to translate the instructions for your platform.

For the most part commands such as `sudo apt install` can be replaced to run on OSX with `brew install`. In those cases where there might be variations, we provide some hints/helpers below. If you run into a specific issue, please create a ticket.


### Install required build dependencies

You will need to install plenty of packages to be able to build Qt. Some of the Qt features are optional, for example support for various databases and if you don't need a specific feature you can skip building the support. Or the other way around, if you need a specific feature you might need to install more packages. See the table below for a list of some optional features and the required development packages you need to install. But first, start by updating your package cache so everything is fresh:

    sudo apt update
    sudo apt-get updatemodesele

    sudo apt install build-essential libfontconfig1-dev libdbus-1-dev libfreetype6-dev libicu-dev libinput-dev libxkbcommon-dev libsqlite3-dev libssl-dev libpng-dev libjpeg-dev libglib2.0-dev libraspberrypi-dev


### Prerequisites

First install all the prerequisites

    # Install Go
    sudo apt install golang
    sudo apt install redis-server redis-tools 

Then start the services that you'll need

    sudo systemctl restart redis.service

Check that redis is running:

    ubuntu@ip-172-31-30-162:~$ redis-cli 
    127.0.0.1:6379> ping
    PONG
    127.0.0.1:6379> 

This project has submodule dependencies, which can be fetched via:

    git submodule init
    git submodule update

### Building and Running

A number of go-related dependcies are necessary to run the project. These should automatically be fetched by `go run`:

    $ go run main.go --help
    go: downloading github.com/go-redis/redis v6.15.8+incompatible
    go: downloading github.com/sirupsen/logrus v1.4.1
    go: downloading github.com/fatih/structs v1.1.0
    Usage of /var/folders/_s/zplk7l31037c_pbcj368ql980000gp/T/go-build654651030/b001/exe/main:
    -redis-addr string
            a string formatted as hostname:port (default "dupl1.local:6379")
    -redis-db int
            an int
    -redis-password string
            a string

You should now be able to test and madke a default build:

    go run main.go --help

To run the application with your local redis instance:

    go run main.go --redis-addr localhost:6379 

and to build a binary

    go build -o ventilator main.go

This will produce a `ventilator` binary, which you can run as:

    ./ventilator --redis-addr localhost:6379

## The GUI

The Python QT GUI for this application is under the gui/ folder, which is loaded as a submodule. Please refer to the [M061-Gui project](https://github.com/dff-labs/M061-gui) for more details. 


