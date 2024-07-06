# JSON Viewer

JSON Viewer is a desktop app for browsing large JSON files. It runs on Linux, Windows and Mac (Experimental).

[![GitHub Release](https://img.shields.io/github/v/release/ErikKalkoken/jsonviewer)](https://github.com/ErikKalkoken/jsonviewer)
[![build status](https://github.com/ErikKalkoken/jsonviewer/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/ErikKalkoken/jsonviewer/actions/workflows/ci-cd.yml)
[![GitHub License](https://img.shields.io/github/license/ErikKalkoken/jsonviewer)](https://github.com/ErikKalkoken/jsonviewer)

## Contents

- [Key Features](#key-features)
- [Screenshots](#screenshots)
- [How to run](#how-to-run)
- [Attributions](#attributions)

## Key Features

- Browse through a JSON document in classic tree structure
- Ability to view large JSON files (>100MB, >1M elements)
- Single executable file, no installation required
- Desktop app that runs on Windows, Linux and macOS (experimental)
- Automatic dark and light mode

## Screenshots

![screenshot](https://cdn.imgpile.com/f/fkuNdSB_xl.png)

## How to run

JSON Viewer is designed to run directly on all platforms without requiring installation. You find the latest packages for all supported platforms on the [releases page](https://github.com/ErikKalkoken/jsonviewer/releases):

- [Linux AppImage](#appimage)
- [Linux TAR file](#tar)
- [Windows](#windows)
- [MAC](#mac-os)
- [Build and run directly](#build-and-run-directly)

### Linux

There are two packages for Linux:

- AppImage
- TAR

We recommend the AppImage package, since it does not require local installation or root privileges to run. The TAR package is there as backup, in case your system can not run AppImages. For more information on the AppImage format please see the [Appimage website](https://appimage.org/).

#### AppImage

Download the latest AppImage file from the releases page and place it a local folder of your choosing (e.g. `~/Applications` or `~/.local/bin`).

Make the AppImage file executable. The execute it ot start the app.

> [!TIP]
> Should you get the following error: `AppImages require FUSE to run.`, you need to first install FUSE on your system. Please see [this page](https://docs.appimage.org/user-guide/troubleshooting/fuse.html#the-appimage-tells-me-it-needs-fuse-to-run) for details.

#### TAR

First download the TAR file from the release page.

Then install the app on your desktop with:

```sh
sudo tar xvfJ evebuddy-v1.0.0-linux-amd64.tar.xz -C /
```

The above command will install the app for all users on your system. User specific data will be stored in the home directories of each user.

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
