# getxpath

[![Build Status](https://travis-ci.com/mat/getxpath.svg?branch=master)](https://travis-ci.com/mat/getxpath)
[![Go Report Card](https://goreportcard.com/badge/github.com/mat/getxpath)](https://goreportcard.com/report/github.com/mat/getxpath)

## Example

<https://getxpath.herokuapp.com/get?url=http://google.com&xpath=//title>

```json
{
	"query": {
		"url": "http://google.com",
		"xpath": "//title"
	},
	"result": "Google",
	"error": null
}
```

## Hosting

An easy way to host this service is to use Heroku, just go to <https://heroku.com/deploy> to get started.

## License

The MIT License (MIT)

Copyright (c) 2015-2019 Matthias Lüdtke, Hamburg - https://github.com/mat

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
