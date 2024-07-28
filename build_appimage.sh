#!/usr/bin/env bash

# This script builds an AppImage from a bundeled Fyne app
# for x86_64 architectures with AppStream metadata.

appname="Janice"
packagename="janice"
appid="io.github.erikkalkoken.janice"
dest="temp.Appdir"
source="temp.Source"

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
./linuxdeploy --appdir "$dest" -v 2 -o appimage -e "$source/usr/local/bin/$packagename"  -d "$source/usr/local/share/applications/$appid.desktop" -i "$source/usr/local/share/pixmaps/$appname.png"

# Cleanup
rm -rf "$source"
rm -rf "$dest"
rm linuxdeploy