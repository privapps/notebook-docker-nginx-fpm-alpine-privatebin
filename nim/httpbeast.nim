import httpbeast, strutils, times, random, os, json, nimcrypto, mimetypes, 
  options, strformat, httpclient, net, async, uri, tables

# Threadvar declarations
var
  dataPath {.threadvar.}: string
  dataPathInitialized {.threadvar.}: bool
  urls {.threadvar.}: seq[string]
  MIMETYPES {.threadvar.}: MimeDB

const REMOTE_TIMEOUT_SECONDS = 10

proc initMimeTypes() =
  MIMETYPES = newMimetypes()
  # Add explicit mappings for common types
  MIMETYPES.register("text/html", "html")
  MIMETYPES.register("text/css", "css")
  MIMETYPES.register("application/javascript", "js")

proc sample[T](s: seq[T]): T =
  if s.len == 0:
    raise newException(ValueError, "Empty sequence passed to sample")
  let idx = rand(s.high)
  return s[idx]

proc getRemoteUrl(): string =
  once:
    let envUrls = getEnv("URLS", "https://paste.rosset.net|https://bin.0xfc.de|https://privatepastebin.com|https://p.darklab.sh")
    urls = envUrls.split('|')
    echo "Initialized URLs: ", urls
  return sample(urls)

proc getDataPath(): string =
  if not dataPathInitialized:
    let envPath = getEnv("NOTE_DATA_PATH", "./notes/")
    dataPath = if envPath.endsWith('/'): envPath else: envPath & "/"
    createDir(dataPath)
    createDir(dataPath & "data")
    dataPathInitialized = true
  return dataPath

proc fileEpoch(fileName: string): int64 =
  try:
    let fileInfo = getFileInfo(fileName)
    return fileInfo.lastWriteTime.toUnix()
  except OSError:
    return 0

proc randomString(length: int): string =
  let byteLen = (length + 1) div 2
  var bytes = newSeq[byte](byteLen)
  for i in 0..<byteLen:
    bytes[i] = byte(rand(255))
  result = bytes.toHex.toLowerAscii()[0..<length]

proc createPwd(): (string, string) =
  var
    pasteid: string
    pwd: string
    filePath: string
  while true:
    pasteid = randomString(6)
    pwd = randomString(6)
    filePath = getDataPath() & pasteid
    if fileEpoch(filePath) == 0:
      break
  writeFile(filePath, pwd)
  setFilePermissions(filePath, {fpUserRead, fpGroupRead, fpOthersRead})
  return (pasteid, pwd)

proc getHost(url: string): string =
  let fp = url.find("//")
  if fp == -1:
    return url
  let remaining = url.substr(fp + 2)
  let ep = remaining.find('/')
  if ep == -1:
    return remaining
  else:
    return remaining[0..<ep]

proc doRelay(url: string, body: string): (string, HttpCode) =
  let host = getHost(url)
  var client = newHttpClient(
    timeout = REMOTE_TIMEOUT_SECONDS * 1000,
    sslContext = newContext(verifyMode = CVerifyNone))
  client.headers = newHttpHeaders({
    "X-Requested-With": "JSONHttpRequest",
    "origin": host
  })
  try:
    let response = client.post(url, body = body)
    let rbody = response.body
    let code = response.code
    if code != Http200:
      return (rbody, code)
    var jsonResult: JsonNode
    try:
      jsonResult = parseJson(rbody)
    except JsonParsingError:
      return (rbody, code)
    if jsonResult.hasKey("status"):
      let status = jsonResult["status"]
      var statusOk = false
      case status.kind
      of JInt: statusOk = status.getInt() == 0
      of JFloat: statusOk = status.getFloat() == 0.0
      else: discard
      if statusOk:
        if jsonResult.hasKey("url"):
          jsonResult["url"] = %(url & $jsonResult["url"])
        else:
          jsonResult["url"] = %url
        return ($jsonResult, Http200)
      else:
        return ("Invalid status", Http500)
    else:
      return ("Missing status", Http500)
  except Exception as e:
    return (e.msg, Http503)

proc getQueryParams(req: Request): Table[string, string] =
  result = initTable[string, string]()
  let fullUri = parseUri(req.path.get())
  for (key, value) in decodeQuery(fullUri.query):
    result[key] = value

proc handleRequest(req: Request) {.async, gcsafe.} =
  let path = req.path.get()
  case req.httpMethod.get()
  of HttpOptions:
    if path == "/relay.php":
      let headers = "Access-Control-Allow-Origin: *\r\n" &
                    "Access-Control-Allow-Methods: POST\r\n" &
                    "Access-Control-Allow-Headers: Content-Type, X-Requested-With"
      req.send(Http200, "", headers)

  of HttpPost:
    if path == "/relay.php":
      let body = req.body.get()
      for i in 0..9:
        let url = getRemoteUrl()
        let (rb, code) = doRelay(url, body)
        if code == Http200:
          let headers = "Content-Type: application/json\r\nAccess-Control-Allow-Origin: *"
          req.send(code, rb, headers)
          echo "Paste remote OK:\t", url
          return
        else:
          echo "Paste remote failed: ", code, " ", url, "\t", rb
      req.send(Http500, "Internal Server Error")

    elif path == "/back.php":
      let (pasteid, pwd) = createPwd()
      let respJson = %*{
        "id": pasteid,
        "key": pwd
      }
      let headers = "Content-Type: application/json"
      req.send(Http200, $respJson, headers)
      echo "Paste registered: ", pasteid

  of HttpGet:
    if path.startsWith("/back.php"):
      let params = req.getQueryParams()
      let pasteid = params.getOrDefault("pasteid")
      if pasteid.len < 4:
        req.send(Http400, "Bad request")
        return
      
      let dataFilePath = getDataPath() & "data/" & pasteid
      if not fileExists(dataFilePath):
        req.send(Http404, "Not Found")
        return
      
      let epoch = fileEpoch(dataFilePath)
      let fileBytes = readFile(dataFilePath)
      let headers = &"Content-Type: application/json\r\nX-Timestamp: {epoch}"
      req.send(Http200, fileBytes, headers)
      echo "Paste served: ", pasteid

    else:
      var filePath = "static" / path
      if dirExists(filePath):
        filePath = filePath / "index.html"
      
      # Normalize path separators
      filePath = filePath.replace("//", "/")
      if fileExists(filePath):
        # Determine MIME type
        let ext = splitFile(filePath).ext.toLowerAscii()
        var mimeType = case ext:
          of ".html": "text/html"
          of ".css": "text/css"
          of ".js": "application/javascript"
          of ".png": "image/png"
          of ".jpg", ".jpeg": "image/jpeg"
          else: MIMETYPES.getMimetype(filePath)
        
        # Fallback for unknown types
        if mimeType == "application/octet-stream":
          mimeType = "text/plain"
        
        let fileContent = readFile(filePath)
        let headers = &"Content-Type: {mimeType}"
        req.send(Http200, fileContent, headers)
      else:
        req.send(Http404, "Not Found")

  of HttpPut:
    if path.startsWith("/back.php"):
      let params = req.getQueryParams()
      
      let pasteid = params.getOrDefault("pasteid")
      if pasteid.len < 4:
        req.send(Http400, "Bad request")
        return
      
      let rHeaders = req.headers.get()
      let xHash = seq[string](rHeaders.getOrDefault("X-Hash", HttpHeaderValues(@[""]))).join("")
      let xTimestamp = seq[string](rHeaders.getOrDefault("X-Timestamp", HttpHeaderValues(@[""]))).join("")
      
      let pwdPath = getDataPath() & pasteid
      if not fileExists(pwdPath):
        req.send(Http404, "Note not found")
        return
      
      let storedKey = readFile(pwdPath)
      let content = req.body.get()
      
      var ctx: sha256
      ctx.init()
      ctx.update(content)
      ctx.update(storedKey)
      var hash: array[32, byte]  # Explicit size for SHA-256 digest
      ctx.finish(hash)
      ctx.clear()
      
      let computedHash = toLowerAscii(toHex(hash))
      if computedHash != xHash:
        echo "Hash mismatch: ", computedHash, " vs ", xHash
        req.send(Http403, "Server token mismatch")
        return
      
      let dataFilePath = getDataPath() & "data/" & pasteid
      let currentEpoch = fileEpoch(dataFilePath)
      var expectedEpoch: int64 = 0
      if xTimestamp.len > 0:
        try:
          expectedEpoch = parseBiggestInt(xTimestamp)
        except ValueError:
          req.send(Http400, "Invalid X-Timestamp")
          return
      if currentEpoch != 0 and currentEpoch != expectedEpoch:
        req.send(Http409, "Conflict")
        return
      
      writeFile(dataFilePath, content)
      let newEpoch = getTime().toUnix()
      let respJson = %*{
        "status": 0,
        "id": pasteid,
        "url": pasteid
      }
      let headers = &"Content-Type: application/json\r\nX-Timestamp: {newEpoch}"
      req.send(Http201, $respJson, headers)
      echo "Paste updated: ", pasteid

    else:
      req.send(Http404, "Not Found")

  of HttpHead:
    req.send(Http200, "")

  else:
    req.send(Http405, "Method Not Allowed")

when isMainModule:
  initMimeTypes()
  let dataDir = getDataPath()
  if not dirExists(dataDir & "data"):
    raise newException(OSError, "ENV NOTE_DATA_PATH not configured right. Missing folder $NOTE_DATA_PATH/data!")
  
  let staticDir = "static"
  if not dirExists(staticDir):
    raise newException(OSError, "Web folder ./static not found!")
  let env_port = getEnv("PORT", "8080")
  let port = Port(env_port.parseInt())
  let settings = initSettings(port = port, numThreads=1)
  echo "Server running on port ", port
  
  let handler: OnRequest = proc (req: Request) {.async, gcsafe, closure.} =
    await handleRequest(req)
  run(handler, settings)