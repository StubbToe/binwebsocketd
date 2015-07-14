#!/usr/bin/nodejs
var binin = 3
var binout = 4;
var fs = require('fs');
var bin = fs.createReadStream(null, {fd: binin, flags: 'r', encoding: 'binary'});
var bout = fs.createWriteStream(null, {fd: binout, flags: 'w', encoding: 'binary'});
var hexChar = ["0", "1", "2", "3", "4", "5", "6", "7","8", "9", "A", "B", "C", "D", "E", "F"];

function byteToHex(b) {
  return hexChar[(b >> 4) & 0x0f] + hexChar[b & 0x0f];
}

function toByteArray(x) {
    var b = [], i = 4;
    while (i--) {
        b[i] = x & 0xFF;
        x=x>>8;
    }
    return b;
}

process.stdin.setEncoding('utf8');
process.stdin.resume();
bin.resume();

process.stdin.on('data', function(data) {
    if (data == null) { return; }
    data = data.substr(0,data.length-1); //trim the \n

    if (data == "sendText") {
        process.stdout.write('Hello World, from the Server\n');
        return;
    } else if (data == "sendBinary") {
        fs.readFile('./static/img.png', 'binary', function(err, str) {
            var buf = new Buffer(str.length + 5);
            buf.writeUInt32LE(str.length, 0);
            buf.write("\x01", 4);
            buf.write(str, 5, 'binary');
            bout.write(buf, 'binary');
        });
        return;
    }

    process.stdout.write('text received: ' + data + "\n");
});

bin.on('data', function(data) {
    var hexPairs = [];
    for (var i = 0; i < data.length; i++) {
        hexPairs[i] = byteToHex(data[i].charCodeAt(0));
    }
    process.stdout.write("binary received: " + hexPairs.join(" ") + "\n" );
    

    var buf = new Buffer(data.length);
    buf.write(data, 0, 'binary');
    bout.write(buf, 'binary');
});

bin.on('error', function(e) {
    process.stdout.write('Error loading binary read stream' + e + "\n");
});

bout.on('error', function(e) {
    process.stdout.write('Error in BINOUT' + e + "\n");
});

