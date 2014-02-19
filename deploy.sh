#!/bin/sh

set -e

git push heroku master
heroku config:set GIT_REVISION=`git describe --always` DEPLOYED_AT=`date +%s`

