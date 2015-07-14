#!/bin/sh

../binwebsocketd --port=8000 --staticdir=./static --dir=./script

# try out the following webpages to see some examples
# http://localhost:8000/index.html     
#     test involving, browser-side: app.js, server-side: process.js
# http://localhost:8000/bintest.html
#     test involving, browser-side: javascript in bintest.html
#     server-side: testscript

