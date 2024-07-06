# JSON Viewer

JSON Viewer is a desktop app for browsing large JSON files. It runs on Linux, Windows 10/11 and Mac OS X (Experimental).

[![GitHub Release](https://img.shields.io/github/v/release/ErikKalkoken/jsonviewer)](https://github.com/ErikKalkoken/jsonviewer)
[![build status](https://github.com/ErikKalkoken/jsonviewer/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/ErikKalkoken/jsonviewer/actions/workflows/ci-cd.yml)
[![GitHub License](https://img.shields.io/github/license/ErikKalkoken/jsonviewer)](https://github.com/ErikKalkoken/jsonviewer)

## Installation

Please follow the instructions below to install JSON viewer on your platform.

### Linux

tbd.

### Windows

tbd.

### MAC

tbd.

### Build and run on-the-fly

If your system is configured to run [Fyne](https://fyne.io/) apps, you can build and run this app directly from the command line:

```sh
go run github.com/ErikKalkoken/jsonviewer@latest
```

For more information on how to configure your system for Fyne please see: [Getting Started](https://docs.fyne.io/started/).

This app is currently under development. Use at your own risk.

## Application performance

JSON viewer is designed to view and browse large JSON files. We found that the key limiting factor for viewing large JSON files is physical RAM. For example on a Linux laptop with 8 GB RAM we were able to view JSON files with ~350 MB size and ~20M elements. Viewing a file of that size consumed ~6.5 GB of RAM, which was about the maximum amount of memory available to applications on that system.

## Attributions

- [Json icons created by LAB Design Studio - Flaticon](https://www.flaticon.com/free-icons/json)
