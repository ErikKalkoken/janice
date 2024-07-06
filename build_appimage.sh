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

# Add AppRun executable
wget https://github.com/AppImage/AppImageKit/releases/download/continuous/AppRun-x86_64 -O "$dest/AppRun"
chmod +x "$dest/AppRun"

# Rearange application files to conform with appdir specs
mkdir "$dest/usr/bin"
mv "$dest/usr/local/share/pixmaps/$appname.png" "$dest"
mv "$dest/usr/local/share/applications/$appname.desktop" "$dest"
mv "$dest/usr/local/bin/$packagename" "$dest/usr/bin"
rm -rf "$dest/usr/local"

# Add category to desktop file
sed -i -- "s/;/\nCategories=$categories;/g" "$dest/$appname.desktop"
desktop-file-validate "$dest/$appname.desktop"

# Create appimage
wget https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage -O appimagetool
chmod +x appimagetool
./appimagetool "$dest"

# Cleanup
rm -rf "$dest"
rm appimagetool