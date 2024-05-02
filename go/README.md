### Run ready to use docker
```
docker run -d -p 8080:8080 --rm ghcr.io/privapps/notebook:v0.0.3
```

### Pre-build Executables

Take a look at github tags top the top, beside the release, you should find the binaries for you platform

#### Directly Run
Your folder structure should look at
```bash
./<executable>
./static/index.html # and more
./notes/data #folder for actual data
```
Then run at the root folder
` env PORT=18080 NOTE_DATA_PATH=./notes/ <executable> `

### Build Go 1.7

```bash
URLS='https://paste.eccologic.net|https://paste.i2pd.xyz|https://pastebin.hot-chilli.net|https://pb.florian2833z.de|https://bin.moritz-fromm.de|https://paste.fizi.ca|https://pastebin.grey.pw|https://paste.tuxcloud.net|https://paste.taiga-san.net|https://vim.cx|https://privatebin.at|https://zerobin.farcy.me|https://snip.dssr.ch|https://bin.snopyta.org|https://paste.danielgorbe.com|https://pastebin.aquilenet.fr|https://pb.nwsec.de|https://wtf.roflcopter.fr/paste|https://paste.systemli.org|https://bin.acquia.com/' go run main.go
```
Need notebook at ./static folder. 

The `URLS` is used for report privatebin publish. 

##### Run inside docker
```
docker run --memory=6m --rm -w -v ${PWD}:/opt  -p 18080:8080 busybox:stable-glibc /opt/<executable>
```

**Note for running inside docker**

busybox:stable-glibc may have problem for publishing https requests to remote Urls. If you need that, try alpine