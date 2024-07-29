#!/usr/bin/env bash

# This script builds an AppImage from a bundeled Fyne app
# for x86_64 architectures with AppStream metadata.

set -e

# Custom variables
buildname="janice"

# Constants
dest="temp.Appdir"
source="temp.Source"

# get tomlq
wget https://github.com/ErikKalkoken/tomlq/releases/download/v0.1.0/tomlq-0.1.0-linux-amd64.tar.gz -O tomlq.tar.gz
tar xf tomlq.tar.gz
rm tomlq.tar.gz

# Use variables from fyne metadata
appname=$(./tomlq -p Details.Name FyneApp.toml)
appid=$(./tomlq -p Details.ID FyneApp.toml)

# Initialize appdir folder
rm -rf "$source"
mkdir "$source"
rm -rf "$dest"
mkdir "$dest"

# Extract application files into appdir folder
tar xvfJ "$appname".tar.xz -C "$source"

# Rename desktop file to match AppStream requirements
mv "$source/usr/local/share/applications/$appname.desktop" "$source/usr/local/share/applications/$appid.desktop"

# Add metadata to AppStream
mkdir -p $dest/usr/share/metainfo
cp "$appid.appdata.xml" "$dest/usr/share/metainfo"

# Create appimage
wget -q https://github.com/linuxdeploy/linuxdeploy/releases/download/continuous/linuxdeploy-x86_64.AppImage -O linuxdeploy
chmod +x linuxdeploy
./linuxdeploy --appdir "$dest" -v 2 -o appimage -e "$source/usr/local/bin/$buildname"  -d "$source/usr/local/share/applications/$appid.desktop" -i "$source/usr/local/share/pixmaps/$appname.png"

# Cleanup
rm -rf "$source"
rm -rf "$dest"
rm linuxdeploy
rm tomlq
