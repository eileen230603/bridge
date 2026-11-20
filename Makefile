# Desarrollo local (macOS/Linux/Windows con make). Requiere Go, Node y Wails v2.

.PHONY: install dev-ap1 dev-ap1-real dev-ap2 build-ap2 build-ap1

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

build-ap1:
	cd apps/ap1-publisher && wails build

build-ap2:
	cd apps/ap2-viewer && wails build
