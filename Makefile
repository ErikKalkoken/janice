bundle:
	fyne bundle --package ui resources > internal/ui/resource.go

appimage:
	./tools/build_appimage.sh

release:
	fyne package -os linux --release