import os
import hashlib
import json
import requests
import random
from microdot import Microdot, Response

# Configuration
EMAIL_SUPPORTED = 0
REMOTE_SUPPORT = 1
DATA_ROOT = os.environ.get('NOTE_DATA_PATH', 'data/')
SERVERS = [
    "https://paste.eccologic.net",
    "https://paste.i2pd.xyz/",
    "https://pastebin.hot-chilli.net/",
    # ... keep all original servers
]
STATIC_DIR =  os.environ.get('NOTE_WEB_PATH', './static')
EMAIL_TEMPLATE = """Please use following URL [...]"""  # Keep original template

app = Microdot()
app.debug = False

class Backend:
    def __init__(self):
        self.conf_dir = DATA_ROOT
        self.data_dir = os.path.join(DATA_ROOT, 'data')
        self._create_dirs()
        
    def _create_dirs(self):
        for d in [self.conf_dir, self.data_dir, STATIC_DIR]:
            if not os.path.exists(d):
                os.mkdir(d)
    
    def register(self, data):
        uid = os.urandom(16).hex()[-6:]
        while uid in os.listdir(self.conf_dir):
            uid = os.urandom(16).hex()[-6:]
        key = hashlib.md5(os.urandom(16)).hexdigest()[-5:]
        
        with open(os.path.join(self.conf_dir, uid), 'w') as f:
            f.write(key)
        
        if EMAIL_SUPPORTED:
            # Email sending placeholder
            pass
            
        return {'id': uid, 'key': key}
    
    def create_or_update(self, pasteid, data, x_hash, timestamp):
        conf_path = os.path.join(self.conf_dir, pasteid)
        if not os.path.exists(conf_path):
            return 404
        
        with open(conf_path, 'r') as f:
            key = f.read().strip()
        
        if x_hash != hashlib.sha256((data + key).encode()).hexdigest():
            return 401
        
        data_path = os.path.join(self.data_dir, pasteid)
        if os.path.exists(data_path):
            current_ts = os.stat(data_path)[8]
            if current_ts != int(timestamp):
                return 409
        
        with open(data_path, 'w') as f:
            f.write(data)
        
        new_ts = os.stat(data_path)[8]
        return (201, new_ts)
    
    def get_paste(self, pasteid):
        data_path = os.path.join(self.data_dir, pasteid)
        if not os.path.exists(data_path):
            return None
        with open(data_path, 'r') as f:
            content = f.read()
        ts = os.stat(data_path)[8]
        return (content, ts)

class RelayWorker:
    def get_server(self):
        return random.choice(SERVERS)
    
    def relay_request(self, data, count=3):
        server = self.get_server()
        try:
            headers = {
                'Content-Type': 'application/json',
                'X-Requested-With': 'JSONHttpRequest',
                'origin': server.split('/')[2]
            }
            
            resp = requests.post(server, data=data, headers=headers)
            json_resp = resp.json()
            
            if json_resp.get('status') == 0:
                json_resp['url'] = server + json_resp['url']
                return json_resp
            elif count > 0:
                return self.relay_request(data, count-1)
            else:
                return {'error': 'Relay failed'}
        except Exception as e:
            return {'error': str(e)}

backend = Backend()
relay_worker = RelayWorker()

@app.route('/back.php', methods=['GET', 'POST', 'PUT', 'OPTIONS'])
def handle_back(request):
    if request.method == 'POST':
        if len(request.body) == 0:
            result = backend.register(request.body)
            return Response(body=json.dumps(result), status_code=201,headers={'Content-Type': 'application/json'})
        else:
            try:
                data = json.loads(request.body)
                if EMAIL_SUPPORTED and ('email' not in data or 'prefix' not in data):
                    return Response(status_code=400)
                
                result = backend.register(data)
                if EMAIL_SUPPORTED:
                    return Response(status_code=201)
                return Response(body=json.dumps(result), headers={'Content-Type': 'application/json'})
            except:
                return Response(status_code=400)
    
    elif request.method == 'PUT':
        pasteid = request.args.get('pasteid')
        x_hash = request.headers.get('X-Hash')
        timestamp = request.headers.get('X-Timestamp', '0')
        
        if not all([pasteid, x_hash, request.body]):
            return Response(status_code=400)
        
        status = backend.create_or_update(pasteid, request.body.decode(), x_hash, timestamp)
        if isinstance(status, tuple):
            return Response(
                body=json.dumps({'status': 0, 'id': pasteid, 'url': pasteid}),
                headers={'Content-Type': 'application/json', 'X-Timestamp': str(status[1])},
                status_code=201
            )
        return Response(status_code=status)
    
    elif request.method == 'GET':
        pasteid = request.args.get('pasteid')
        if not pasteid:
            return Response(status_code=400)
        
        result = backend.get_paste(pasteid)
        if result:
            return Response(
                body=result[0],
                headers={'Content-Type': 'application/json', 'X-Timestamp': str(result[1])}
            )
        return Response(status_code=404)
    
    elif request.method == 'OPTIONS':
        return Response(headers={
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Methods': 'GET,POST,PUT,OPTIONS',
            'Access-Control-Allow-Headers': 'Content-Type, X-Hash, X-Timestamp'
        })
    
    return Response(status_code=405)

@app.route('/relay.php', methods=['POST', 'OPTIONS'])
def handle_relay(request):
    if request.method == 'OPTIONS':
        return Response(headers={
            'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Methods': 'POST',
            'Access-Control-Allow-Headers': 'Content-Type, X-Requested-With'
        })
    
    try:
        data = json.loads(request.body)
        if not all(k in data for k in ('v', 'ct', 'adata')):
            return Response(status_code=400)
        
        result = relay_worker.relay_request(request.body)
        return Response(
            body=json.dumps(result),
            headers={'Content-Type': 'application/json'},
            status_code=201
        )
    except:
        return Response(status_code=400)

@app.route('/<path:path>')
def serve_static(request, path):
    safe_path = path.lstrip('/')
    full_path = os.path.join(STATIC_DIR, safe_path)
    
    if '..' in full_path or not os.path.exists(full_path):
        return Response(status_code=404)
    if os.path.isdir(full_path):
        full_path = os.path.join(full_path, 'index.html')
        if not os.path.exists(full_path):
            return Response(status_code=404)
    
    return Response.send_file(full_path)

@app.route('/')
def serve_root(request):
    return serve_static(request, '/index.html')

@app.after_request
def add_cors_headers(request, response):
    if REMOTE_SUPPORT:
        response.headers['Access-Control-Allow-Origin'] = '*'
    return response

if __name__ == '__main__':
    port = os.environ.get('PORT', '8080')
    print('listening: ', port)
    app.run(port=int(port))
