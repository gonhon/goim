# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

all: test build
build:
	rm -rf target/
	mkdir target/
	cp cmd/comet/comet-example.toml target/comet.toml
	cp cmd/logic/logic-example.toml target/logic.toml
	cp cmd/job/job-example.toml target/job.toml
	$(GOBUILD) -gcflags="all=-N -l" -o target/comet cmd/comet/main.go
	$(GOBUILD) -gcflags="all=-N -l" -o target/logic cmd/logic/main.go
	$(GOBUILD) -gcflags="all=-N -l" -o target/job cmd/job/main.go

test:
	$(GOTEST) -v ./...

clean:
	rm -rf target/

run:
	nohup target/logic -conf=target/logic.toml -region=sh -zone=sh001 -deploy.env=dev -weight=10  -logtostderr=true -v=2 > target/logic.log 2>&1 &
	nohup target/comet -conf=target/comet.toml -region=sh -zone=sh001 -deploy.env=dev -weight=10 -addrs=127.0.0.1 -debug=true  -logtostderr=true -v=2 > target/comet.log 2>&1 &
	nohup target/job -conf=target/job.toml -region=sh -zone=sh001 -deploy.env=dev  -logtostderr=true -v=2 > target/job.log 2>&1 &

stop:
	pkill -f target/logic
	pkill -f target/job
	pkill -f target/comet
