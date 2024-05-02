# Private notebook back end and docker images

There are three kind of backend
* Go backend, which can edit existing noted
* PHP backend, which can edit existing noted
* PHP backend, privatebin which is unmodifiable. (only client side configuration is sufficient)

Currently we have Go backend and PHP privatebin backend dockers.

# End user note

## Custom Backend Support Edit Exiting Note
```
#mkdir notes; chmod a+w notes
docker run --memory=6m --rm -p 18080:8080 -v $PWD/notes:/notes ghcr.io/privapps/notebook:v0.0.3
```

After you finish your note for the first version, go to `Settings => Hosted:Editable` then register and generate a `Symmetric Key`.

Save the new generated link and you can update the exiting node later.

To see detail information about go executable command, see [go](./go/)

Binaries can be found at tags area. And you might use different configuration. See [php/config.json](php/config.json) and [config.json](config.json)


## Docker Images for Privatebin

It contains privapps notebook, privatebin, nginx, php-fpm & alpine 

*This image is based on the docker image of [privatebin](https://github.com/PrivateBin/docker-nginx-fpm-alpine)*

It adds the front end UI, and use privatebin as the back end storage. Note this images is at docker hub and only have amd64.

To run it 
```
#mkdir privatebin-data; chmod a+w privatebin-data
docker run -d --restart="always" --read-only -p 8080:8080 -v $PWD/privatebin-data:/srv/data docker.io/privapps/notebook
```
More details please take a look at https://github.com/PrivateBin/docker-nginx-fpm-alpine

# Developers Note
There is a custom build backend server that you can modified existing saved notes, other than use privatebin, you can get the go binaries at release.

#### Directly Run
Your folder structure should look at
```bash
./<executable>
./static/index.html # and more
./notes/data        #folder for actual data. back it up if you need to
```

Then run at the root folder
` env PORT=18080 NOTE_DATA_PATH=./notes/ <executable> `

Or Run inside docker
`docker run --rm -w /var/www --entrypoint /var/www/<executable> -v ${PWD}:/var/www  -p 18080:8080 -e NOTE_DATA_PATH=./notes/ alpine:3.13`
