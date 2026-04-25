.PHONY: build web-build embed test lint smoke fixture clean

GO_BIN := chop
EMBED_DIST := internal/dashboard/dist
FIXTURE := testdata/pattern.mp4

build: embed
	go build -o $(GO_BIN) ./cmd/chop

web-build:
	cd web && pnpm install --frozen-lockfile && pnpm build

embed: web-build
	cp -r web/dist/* $(EMBED_DIST)/

test:
	go test -race ./...

lint:
	golangci-lint run

$(FIXTURE):
	ffmpeg -y -f lavfi -i testsrc2=size=854x480:rate=30 -t 5 \
		-c:v libx264 -profile:v baseline -tune zerolatency \
		-g 30 -pix_fmt yuv420p -movflags +faststart -an $@

fixture: $(FIXTURE)

smoke: build fixture
	./scripts/smoke.sh

clean:
	rm -rf $(GO_BIN) web/dist coverage.txt $(EMBED_DIST)
	mkdir -p $(EMBED_DIST)
	touch $(EMBED_DIST)/.gitkeep
