var hexChar = ["0", "1", "2", "3", "4", "5", "6", "7","8", "9", "A", "B", "C", "D", "E", "F"];
function byteToHex(b) {
  return hexChar[(b >> 4) & 0x0f] + hexChar[b & 0x0f];
}


function log(msg) {
    $('#log').append('<div>' + msg + '</div>');
}

var ws = new WebSocket('ws://pounce.house:8080/');
ws.onopen = function() {
    log('CONNECT');
};
ws.onclose = function() {
    log('DISCONNECT');
};
ws.onmessage = function(event) {
    var b = toBytes(event.data);

    if (b[0] == 0) {
        readText(event.data.substr(1));
    } else {
        readBinary(event.data);
    }

    log('MESSAGE: ' + event.data);
};

function toBytes(str) {
    var bytes = [];

    for (var i = 0; i < str.length; ++i) {
        bytes.push(str.charCodeAt(i));
    }

    return bytes;
};

function readBinary(data) {
    var hexPairs = [];
    var reader = new FileReader();
    reader.addEventListener("loadend", function() {
        b = new Uint8Array(reader.result);
        for (var i =1; i < b.length; i++) {
            hexPairs.push(byteToHex(b[i]));
        }
        $('#getBinaryData').val(hexPairs.join(" "));
    });
    reader.readAsArrayBuffer(data);
};

function readText(data) {
    $('#getTextData').val(data);
};

function sendBinary(ws, type, data) {
    ws.send(String.fromCharCode(type) + data);
};

function sendText(ws, data) {
    ws.send(String.fromCharCode(0) + data);
};

$(document).ready(function() {
    $('#sendTextButton').on('click', function(e) {
        e.preventDefault();
        sendText(ws, $('#sendTextData').val());
    });

    $('#sendBinaryButton').on('click', function(e) {
        e.preventDefault();
        sendBinary(ws, parseInt($('#sendBinaryType').val()), $('#sendBinaryData').val());
    });

    $('#getTextButton').on('click', function(e) {
        e.preventDefault();
        sendText(ws, 'sendText');
    });

    $('#getBinaryButton').on('click', function(e) {
        e.preventDefault();
        sendText(ws, 'sendBinary'); 
    });
});
