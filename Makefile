# Desarrollo local (macOS/Linux/Windows con make). Requiere Go, Node y Wails v2.

VERSION := $(shell cat VERSION)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
DATE    := $(shell date -u +%Y-%m-%d)
LDFLAGS := -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(DATE)

.PHONY: install dev-ap1 dev-ap1-real dev-ap2 build-ap2 build-ap1 version version-check

install:
	npm --prefix apps/ap1-publisher/frontend install
	npm --prefix apps/ap2-viewer/frontend install

# AP1 con datos dummy: sin Symphony, sin Epson y con licencia simulada.
# Todo se guarda en runtime/dev (ignorado por git).
dev-ap1:
	cd apps/ap1-publisher && AP1_DEV=1 wails dev

# AP1 contra servicios reales (config.json)
dev-ap1-real:
	cd apps/ap1-publisher && wails dev

# AP2: Viewer. Usa runtime/temp/<StudyUID> como contenido si se pasa STUDY=<uid>
dev-ap2:
	cd apps/ap2-viewer && $(if $(STUDY),DICOM_VIEWER_CONTENT_DIR=$(CURDIR)/runtime/temp/$(STUDY)) wails dev

# Muestra la versión vigente (fuente: archivo VERSION)
version:
	@echo $(VERSION)

# Falla si VERSION, version.go y wails.json no coinciden
version-check:
	cd apps/ap1-publisher && go test -run TestVersion .

build-ap1: version-check
	cd apps/ap1-publisher && wails build -ldflags "$(LDFLAGS)"

build-ap2:
	cd apps/ap2-viewer && wails build
