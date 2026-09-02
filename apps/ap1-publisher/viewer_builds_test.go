package main

import (
	"io/fs"
	"reflect"
	"testing"
)

func TestEmbeddedViewerBuildInventory(t *testing.T) {
	var files []string
	err := fs.WalkDir(viewerBuilds, "viewer-builds", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"viewer-builds/macos/Symphony Viewer.app/Contents/Info.plist",
		"viewer-builds/macos/Symphony Viewer.app/Contents/MacOS/Portable DICOM Viewer",
		"viewer-builds/macos/Symphony Viewer.app/Contents/Resources/iconfile.icns",
		"viewer-builds/macos/Symphony Viewer.app/Contents/_CodeSignature/CodeResources",
		"viewer-builds/windows/Symphony Viewer.exe",
	}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("unexpected embedded viewer inventory:\n got: %#v\nwant: %#v", files, want)
	}
}
