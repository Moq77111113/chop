.PHONY: build test lint smoke fixture clean

GO_BIN := chop
FIXTURE := testdata/pattern.mp4

build:
	go build -o $(GO_BIN) ./cmd/chop

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
	rm -rf $(GO_BIN) coverage.txt
