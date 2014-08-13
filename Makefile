
update_godeps: 
	godep save .

deploy:
	git push heroku master
	heroku config:set GIT_REVISION=`git describe --always` DEPLOYED_AT=`date +%s`

run_server:
	go run xpathfetcher.go -port=3000
