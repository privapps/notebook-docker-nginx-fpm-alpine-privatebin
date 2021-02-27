## Docker Images for Privapp Notebook

It contains privapps notebook, privatebin, nginx, php-fpm & alpine 

##### This image is based on the docker image of [privatebin](https://github.com/PrivateBin/docker-nginx-fpm-alpine)

It adds the front end UI, and use privatebin as the back end storage.

To run it 
```
docker run -d --restart="always" --read-only -p 8080:8080 -v $PWD/privatebin-data:/srv/data privapps/notebook
```
More details please take a look at https://github.com/PrivateBin/docker-nginx-fpm-alpine

