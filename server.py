# Source code modified by: Jan Tache
#
# Base source code:
# https://gist.github.com/mafayaz/faf938a896357c3a4c9d6da27edcff08
#
# This file takes information from HTTP POST from the ThingPark
# server and logs it into human-readable format
#
# Usage: python <filename> <port number> <IP address>

from BaseHTTPServer import BaseHTTPRequestHandler,HTTPServer
from SocketServer import ThreadingMixIn
import threading
import argparse
import re
import cgi
import json

raw_out_file = '/home/jan/server_output.txt'
parsed_out_file = '/home/jan/parsed_lora_info.txt'

class HTTPRequestHandler(BaseHTTPRequestHandler):
    # Code to handle HTTP POST
    def do_GET(s):
        if s.path == '/svg.svg':
            with open('svg.svg', 'r') as f:
                s.wfile.write(f.read())
            return

        s.send_response(200)
        s.send_header("Content-type", "text/html")
        s.end_headers()
        s.wfile.write("\n\
<html><head><title>This is the worst hack-job I've ever done</title></head><body><p>Something something svg...</p>\n\
<object id='roo' data='svg.svg' type='image/svg+xml'></object><script> \n\
function getSvg() { \n\
console.log('hello')\n\
var oldObj = document.getElementById('roo')\n\
var newObj = document.createElement('object')\n\
newObj.data = 'svg.svg'\n\
newObj.type = 'image/svg+xml'\n\
newObj.id = 'roo'\n\
oldObj.remove()\n\
document.body.appendChild(newObj)\n\
} \n\
 \n\
function draw() { \n\
	var j = 50 \n\
 \n\
	for (var a= [], i = 0; i < 8192; i++) { \n\
		a[i] = i % j ? 0 : j \n\
	} \n\
	//var last = document.body.appendChild(getSvg()) \n\
 \n\
	var f = function() { \n\
                getSvg()\n\
	} \n\
 \n\
	setInterval(f, 1000) \n\
} \n\
 \n\
draw()\n\
 </script></body></html> ")

# Rest of source code was unmodified from the original; setting up HTTP server
class ThreadedHTTPServer(ThreadingMixIn, HTTPServer):
    allow_reuse_address = True

    def shutdown(self):
        self.socket.close()
        HTTPServer.shutdown(self)

class SimpleHttpServer():
    def __init__(self, ip, port):
        self.server = ThreadedHTTPServer((ip,port), HTTPRequestHandler)

    def start(self):
        self.server_thread = threading.Thread(target=self.server.serve_forever)
        self.server_thread.daemon = True
        self.server_thread.start()

    def waitForThread(self):
        self.server_thread.join()

    def addRecord(self, recordID, jsonEncodedRecord):
        LocalData.records[recordID] = jsonEncodedRecord

    def stop(self):
        self.server.shutdown()
        self.waitForThread()

if __name__=='__main__':
    parser = argparse.ArgumentParser(description='HTTP Server')
    parser.add_argument('port', type=int, help='Listening port for HTTP Server')
    parser.add_argument('ip', help='HTTP Server IP')
    args = parser.parse_args()

    server = SimpleHttpServer(args.ip, args.port)
    print 'HTTP Server Running...........'
    server.start()
    server.waitForThread()
