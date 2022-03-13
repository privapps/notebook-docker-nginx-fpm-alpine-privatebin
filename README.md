## Custom Backend Support Edit Exiting Note

### Docker Images
```
docker run --memory=6m --rm -p 18080:8000 ghcr.io/privapps/notebook:dev
```

To see detail information about go executable command, see branch [go](./go/)

Binaries can be found at tags area. And you might use different configuration. See [php/config.json](php/config.json)


## Docker Images for Privapp Notebook

It contains privapps notebook, privatebin, nginx, php-fpm & alpine 

##### This image is based on the docker image of [privatebin](https://github.com/PrivateBin/docker-nginx-fpm-alpine)

It adds the front end UI, and use privatebin as the back end storage.

To run it 
```
docker run -d --restart="always" --read-only -p 8080:8080 -v $PWD/privatebin-data:/srv/data privapps/notebook
```
More details please take a look at https://github.com/PrivateBin/docker-nginx-fpm-alpine

## Save-able Backend
There is a custom build backend server that you can modified existing saved notes, other than use privatebin.

After you finish your note for the first version, go to `Settings => Hosted:Editable` then register and generate a `Symmetric Key`.

Save the new generated link and you can update the exiting node.

Source code of it is at `relays` branch, it has implementation for both Golang and PHP.

### Pre-build Go Executables

Take a look at branch `latest-binaries` you should find the go binary for you platform. The use the web there as it is pre-configured.

##### Directly Run
Your folder structure should look at
```bash
./<executable>
./static/index.html # and more
./notes/data        #folder for actual data. back it up if you need to
```
Then run at the root folder
` env PORT=18080 NOTE_DATA_PATH=./notes/ <executable>

##### Run inside docker
```
docker run --rm -w /var/www --entrypoint /var/www/<executable> -v ${PWD}:/var/www  -p 18080:8000 -e NOTE_DATA_PATH=./notes/ alpine:3.13
```
