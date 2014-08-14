
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
	go get code.google.com/p/go.tools/cmd/vet
	go get github.com/golang/lint/golint

check:
	go vet *.go
	golint *.go

