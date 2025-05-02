#!/usr/bin/env bash

# This script builds an AppImage with AppStream metadata from a Fyne app

set -e

# Constants
dest="temp.Appdir"
source="temp.Source"

# get fynemeta
wget https://github.com/ErikKalkoken/fynemeta/releases/download/v0.1.1/fynemeta-0.1.1-linux-amd64.tar.gz -O fynemeta.tar.gz
tar xf fynemeta.tar.gz
rm fynemeta.tar.gz

# Use variables from fyne metadata
appname=$(./fynemeta lookup -k Details.Name FyneApp.toml)
appid=$(./fynemeta lookup -k Details.ID FyneApp.toml)
buildname=$(./fynemeta lookup -k Release.BuildName FyneApp.toml)

# Initialize appdir folder
rm -rf "$source"
mkdir "$source"
rm -rf "$dest"
mkdir "$dest"

# Extract application files into appdir folder
tar xvfJ "$appname".tar.xz -C "$source"

# Rename desktop file to match AppStream requirements
mv "$source/usr/local/share/applications/$appname.desktop" "$source/usr/local/share/applications/$appid.desktop"

# Add AppStream appdata file
mkdir -p $dest/usr/share/metainfo
./fynemeta generate -t AppStream -d "$dest/usr/share/metainfo"

# Create appimage
wget -q https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage -O linuxdeploy
chmod +x linuxdeploy
./linuxdeploy --appdir "$dest" -v 2 -o appimage -e "$source/usr/local/bin/$buildname"  -d "$source/usr/local/share/applications/$appid.desktop" -i "$source/usr/local/share/pixmaps/$appname.png"

# Cleanup
rm -rf "$source"
rm -rf "$dest"
rm linuxdeploy
rm fynemeta
