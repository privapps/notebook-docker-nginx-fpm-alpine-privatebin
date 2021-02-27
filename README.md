## Docker Images for Privapp Notebook used by cloud foundry

It contains privapps notebook, privatebin, nginx, php-fpm & alpine 

In IBM Cloud Foundry, app needs to disable IPv6 and run with super user.
```
ibmcloud cf push <app-name> --docker-image privapps/notebook-cf
```

If you need run as regular docker or kubernetes, check the main branch.
