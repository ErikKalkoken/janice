# JSON Viewer

A desktop app for viewing large JSON files

[![GitHub Release](https://img.shields.io/github/v/release/ErikKalkoken/jsonviewer)](https://github.com/ErikKalkoken/jsonviewer)
[![build status](https://github.com/ErikKalkoken/jsonviewer/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/ErikKalkoken/jsonviewer/actions/workflows/ci-cd.yml)
[![GitHub License](https://img.shields.io/github/license/ErikKalkoken/jsonviewer)](https://github.com/ErikKalkoken/jsonviewer)

## Contents

- [Description](#description)
- [Screenshots](#screenshots)
- [How to run](#how-to-run)
- [Attributions](#attributions)

## Description

JSON Viewer is a desktop all for viewing large JSON files. It's key features are:

- Browse through a JSON document in classic tree structure
- JSON files can be opened via file dialog, from clipboard or drag-and-drop
- Supports viewing very large JSON files (>100MB, >10M elements)
- Search for keys in the document. Supports wildcards.
- Export segments of a JSON file into a new file
- Copy values to clipboard
- Single executable file, no installation required
- Desktop app that runs on Windows, Linux and macOS (experimental)
- Automatic dark and light mode

## Screenshots

![screenshot](https://cdn.imgpile.com/f/Nqv5wTv_xl.png)

## How to run

JSON Viewer is shipped as a single executable and designed to run without requiring any installation. You find the latest packages for download on the [releases page](https://github.com/ErikKalkoken/jsonviewer/releases).

### Linux

The app is shipped in the [AppImage](https://appimage.org/) format, so it can be used without requiring installation and on many different Linux distributions. Download the latest AppImage file from the releases page and make it executable. Then execute it to start the app.

> [!TIP]
> Should you get the following error: `AppImages require FUSE to run.`, you need to first install FUSE on your system. Thi s is a library required by all AppImages to function. Please see [this page](https://docs.appimage.org/user-guide/troubleshooting/fuse.html#the-appimage-tells-me-it-needs-fuse-to-run) for details.

### Windows

First download the windows zip file from the latest release on Github.

Then unzip the file into a directory of your choice and run the .exe file to start the app.

### Mac OS

> [!NOTE]
> The MAC version is currently experimental only, since we have not been able to verify that the release process actually works. We would very much appreciate any feedback on wether the package works or what needs to be improved.

First download the darwin zip file from the latest release on Github.

Then unzip the file into a directory of your choice and run the .app file to start the app.

### Build and run directly

If your system is configured to build [Fyne](https://fyne.io/) apps, you can build and run this app directly from the repository with the following command:

```sh
go run github.com/ErikKalkoken/jsonviewer@latest
```

For more information on how to configure your system for Fyne please see: [Getting Started](https://docs.fyne.io/started/).

## Attributions

- [Json icons created by LAB Design Studio - Flaticon](https://www.flaticon.com/free-icons/json)
