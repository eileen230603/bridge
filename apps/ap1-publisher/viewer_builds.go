package main

import "embed"

// viewerBuilds contains only the prebuilt AP2 applications. Patient manifests
// and DICOM data are generated separately for every study package.
//
//go:embed viewer-builds/*
var viewerBuilds embed.FS
