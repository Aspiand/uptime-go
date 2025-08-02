from http.server import BaseHTTPRequestHandler, HTTPServer
from time import sleep
import json

class SimpleHTTPRequestHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path == '/uptime/up' or self.path == '/uptime/down':
            content_length = int(self.headers['Content-Length'])
            body = self.rfile.read(content_length)
            print(f"Received POST request on {self.path} with body: {body.decode('utf-8')}")
            self.send_response(200)
            self.end_headers()
            self.wfile.write(b'Request received')            
        else:
            self.send_response(404)
            self.end_headers()
            self.wfile.write(b'Not Found')

    def do_GET(self):
        sleep(60)
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b'Request received')

def run(server_class=HTTPServer, handler_class=SimpleHTTPRequestHandler, port=8005):
    server_address = ('', port)
    httpd = server_class(server_address, handler_class)
    print(f"Starting httpd on port {port}...")
    httpd.serve_forever()

if __name__ == "__main__":
    run()
