Go 1.7

```bash
URLS='https://paste.eccologic.net|https://paste.i2pd.xyz|https://pastebin.hot-chilli.net|https://pb.florian2833z.de|https://bin.moritz-fromm.de|https://paste.fizi.ca|https://pastebin.grey.pw|https://paste.tuxcloud.net|https://paste.taiga-san.net|https://vim.cx|https://privatebin.at|https://zerobin.farcy.me|https://snip.dssr.ch|https://bin.snopyta.org|https://paste.danielgorbe.com|https://pastebin.aquilenet.fr|https://pb.nwsec.de|https://wtf.roflcopter.fr/paste|https://paste.systemli.org|https://bin.acquia.com/' go run main.go
```
Need notebook at ./static folder. can run with memory as low as 8mb.

### Pre-build Executables

Take a look at branch `latest-binaries` you should find the binary for you platform

##### Directly Run
Your folder structure should look at
```bash
./<executable>
./static/index.html # and more
./notes/data #folder for actual data
```
Then run at the root folder
` env PORT=18080 NOTE_DATA_PATH=./notes/ <executable> `

##### Run inside docker
```
docker run --rm -w /var/www --entrypoint /var/www/<executable> -v ${PWD}:/var/www  -p 18080:8000 -e NOTE_DATA_PATH=./notes/ alpine:3.13
```
