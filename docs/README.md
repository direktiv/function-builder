 # Direktiv Service Builder

Creating a new service takes three steps: configuring the input, the command, and output. This documentation will go through these three stages in detail, as well as all configuration choices. As a good technical individual, you may believe it's easier to just skip the documentation and go straight to the examples(examples/README.md) as soon as possible. :wink:

- [Initializing the Service](#initializing-the-service)
- [Configuring the Input](#configuring-the-input)
    - [Adding Commands](#adding-commands)
    - [Adding Foreach](#adding-foreach)
    - [Adding HTTP Requests](#adding-http-requests)
    - [Adding HTTP Foreach](#adding-http-foreach)
- [Configuring the Output](#configuring-the-output)
- [Advanced Features](#advanced-features)

## Initializing the Service

The first step starting with a new service is initializing the project. Direktiv's service builder comes with a docker container so no local installation is required. To initializ a project simply call the following command:

```
docker run -v `pwd`:/tmp/app direktiv/action-builder init myservice
```

The container maps a local directory into the container and uses this directory as the base. In the above exmaple we are using the Linux `pwd` command but this can be a static file path as well. The last argument is the name of the service.


> **TIP**: On Linux the owner of the created files is `root`. To avoid this pass in a user and group.

*Passing user and group id to the container*

```
docker run --user 1000:1000 -v `pwd`:/tmp/app direktiv/action-builder init myservice
```




<!-- - using output from former command
- if statements in print
- print docs -->