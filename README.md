# Janice

A desktop app for viewing large JSON files.

[![GitHub Release](https://img.shields.io/github/v/release/ErikKalkoken/janice)](https://github.com/ErikKalkoken/janice)
[![build status](https://github.com/ErikKalkoken/janice/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/ErikKalkoken/janice/actions/workflows/ci-cd.yml)
[![GitHub License](https://img.shields.io/github/license/ErikKalkoken/janice)](https://github.com/ErikKalkoken/janice)

## Contents

- [Description](#description)
- [Screenshots](#screenshots)
- [How to run](#how-to-run)
- [FAQ](#faq)
- [Attributions](#attributions)

## Description

Janice is a desktop app for viewing large JSON files. It's key features are:

- Browse through a JSON document in classic tree structure
- JSON files can be opened via file dialog, from clipboard, dropped on the window or given as command line argument
- Supports viewing very large JSON files (>100MB, >10M elements)
- Search for keys and values in the document. Supports wildcards.
- Export parts of a JSON file into a new file or to clipboard
- Copy values to clipboard
- Single executable file, no installation required
- Desktop app that runs on Windows, Linux and macOS (experimental)
- Automatic dark and light theme

## Screenshots

### Light theme

![light](https://cdn.imgpile.com/f/0IrYBjJ_xl.png)

### Dark theme

![dark](https://cdn.imgpile.com/f/bdQBc3q_xl.png)

## How to run

To run Janice just download and unzip the latest release to your computer. Janice ships as a single executable file that can be run directly. You find the latest packages for download on the [releases page](https://github.com/ErikKalkoken/janice/releases).

### Linux

> [!NOTE]
> The app is shipped in the [AppImage](https://appimage.org/) format, so it can be used without requiring installation and run on many different Linux distributions.

1. Download the latest AppImage file from the releases page and make it executable.
1. Execute it to start the app.

> [!TIP]
> Should you get the following error: `AppImages require FUSE to run.`, you need to first install FUSE on your system. Thi s is a library required by all AppImages to function. Please see [this page](https://docs.appimage.org/user-guide/troubleshooting/fuse.html#the-appimage-tells-me-it-needs-fuse-to-run) for details.

### Windows

1. Download the windows zip file from the latest release on Github.
1. Unzip the file into a directory of your choice and run the .exe file to start the app.

### Mac OS

> [!NOTE]
> The MAC version is currently experimental only, since we have not been able to verify that the release process actually works. We would very much appreciate any feedback on wether the package works or what needs to be improved.

1. Download the darwin zip file from the latest release on Github.
1. Unzip the file into a directory of your choice
1. Run the .app file to start the app.

> [!TIP]
> MacOS may report this app incorrectly as "damaged", because it is not signed with an Apple certificate. You can remove this error by opening a terminal and running the following command. For more information please see [Fyne Troubleshooting](https://docs.fyne.io/faq/troubleshoot#distribution):
>
> ```sudo xattr -r -d com.apple.quarantine Janice.app```

### Build and run from repository

If your system is configured to build [Fyne](https://fyne.io/) apps, you can build and run this app directly from the repository with the following command:

```sh
go run github.com/ErikKalkoken/janice@latest
```

For more information on how to configure your system for Fyne please see: [Getting Started](https://docs.fyne.io/started/).

## FAQ

### What is the largest JSON file that I can load?

The largest JSON file you can load on your computer depends mainly on how much RAM you have and on the particular JSON file. The main driver for memory consumption is the number of elements in a JSON document.

For comparison we did a load test on one of our developer notebooks. It has 8 GB RAM and runs Ubuntu 22.04 LTS. We were able to load a JSON files successfully with up to 45 million elements. The size of our test file was about 2.5 GB.

## Are JSON files formatted?

Yes. The JSON document is rendered as tree and keys are shown in alphabetical order.

## Attributions

- [Json icons created by LAB Design Studio - Flaticon](https://www.flaticon.com/free-icons/json)
