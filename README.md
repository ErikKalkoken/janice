# JSON Viewer

This is a desktop app to view large JSON file. It runs on Windows 10/11, Mac OS X and Linux.

If your system is configured to run [Fyne](https://fyne.io/) apps, you can start this app directly from the command line with:

```sh
go run github.com/ErikKalkoken/jsonviewer@latest
```

For more information on how to configure your system for Fyne please see: [Getting Started](https://docs.fyne.io/started/).

This app is currently under development. Use at your own risk.

## Performance

We tested JSON Viewer with large files. Here is what we found so far:

- 350MB with 20M elements => works and uses about 6.5 GB RAM

## Attributions

- [Json icons created by LAB Design Studio - Flaticon](https://www.flaticon.com/free-icons/json)
