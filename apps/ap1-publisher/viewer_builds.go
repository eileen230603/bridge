package main

import "embed"

//go:embed config.json
var defaultConfig []byte
//go:embed frontend/src/assets/SYMPHONYPNG.png
var symphonyLogo []byte
//go:embed frontend/src/assets/MEDIGLOBEPNG.png
var mediglobeLogo []byte
// viewerBuilds contains only the prebuilt AP2 applications. Patient manifests
// and DICOM data are generated separately for every study package.
//
//go:embed "viewer-builds/windows/Symphony Viewer.exe" "all:viewer-builds/macos/Symphony Viewer.app"
var viewerBuilds embed.FS
