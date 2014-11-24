
test:
	go test

update_godeps: 
	godep save .

deploy:
	git push heroku master
	heroku config:set GIT_REVISION=`git describe --always` DEPLOYED_AT=`date +%s`

run_server:
	go run getxpath.go -port=3000

install_devtools:
	go get code.google.com/p/go.tools/cmd/cover
	go get code.google.com/p/go.tools/cmd/godoc
	go get code.google.com/p/go.tools/cmd/vet
	go get github.com/golang/lint/golint
	go get github.com/tools/godep

check:
	go vet *.go
	golint *.go

vendor_dependencies:
	godep save -r ./...
	# Need to go get in order to fill $GOPATH/pkg... to minimize compile times.
	go get github.com/mat/getxpath/Godeps/_workspace/src/code.google.com/p/go.net/html/charset
	go get github.com/mat/getxpath/Godeps/_workspace/src/github.com/moovweb/gokogiri

