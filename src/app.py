import http.server
import socketserver
import os

# HTTP port
PORT = os.environ.get("PORT", 8080)
# Configmap path inside container
CONFIGMAP_FILE_PATH = os.environ.get("CONFIGMAP_FILE_PATH", "/app/config/config.txt")

# Class to handle HTTP requests
class MyRequestHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/':
            configmap_content = self.get_configmap_content()
            if configmap_content:
                self.send_response(200)
                self.send_header("Content-type", "application/json")
                self.end_headers()
                self.wfile.write(configmap_content.encode())
            else:
                self.send_response(500)
                self.send_header("Content-type", "text/plain")
                self.end_headers()
                self.wfile.write("Error retrieving ConfigMap content".encode())
        else:
            print(f"Path not found: {self.path}")
            self.send_response(404)
            self.send_header("Content-type", "text/plain")
            self.end_headers()
            self.wfile.write("Not Found".encode())

    def get_configmap_content(self):
        try:
            with open(CONFIGMAP_FILE_PATH, 'r') as configmap_file:
                configmap_content = configmap_file.read()
            return configmap_content
        except Exception:
            return None

# Configure the server with this custom request handler
with socketserver.TCPServer(("", PORT), MyRequestHandler) as httpd:
    print(f"Serving at port {PORT}")
    httpd.serve_forever()
