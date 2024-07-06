#!/usr/bin/env bash

# This script builds an AppImage from a bundeled Fyne app for x86_64 architectures

appname="JSON Viewer"
packagename="jsonviewer"
categories="Utility"
dest="temp.Appdir"

# Initialize appdir folder
rm -rf "$dest"
mkdir "$dest"

# Extract application files into appdir folder
tar xvfJ "$appname".tar.xz -C "$dest"

# Add category to desktop file
sed -i -- "s/;/\nCategories=$categories;/g" "$dest/usr/local/share/applications/$appname.desktop"
# desktop-file-validate "$dest/$appname.desktop"

# Create appimage
wget https://github.com/probonopd/linuxdeployqt/releases/download/continuous/linuxdeployqt-continuous-x86_64.AppImage -O linuxdeployqt
chmod +x linuxdeployqt
./linuxdeploy --appdir "$dest" -o appimage -e "$dest/usr/local/bin/$packagename"  -d "$dest/usr/local/share/applications/$appname.desktop" -i "$dest/usr/local/share/pixmaps/$appname.png"

# Cleanup
rm -rf "$dest"
# rm appimagetool