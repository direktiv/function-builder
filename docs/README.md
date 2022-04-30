 # Direktiv Service Builder

Creating a new service takes three steps: configuring the input, the command, and output. This documentation will go through these three stages in detail, as well as all configuration choices. As a good technical individual, you may believe it's easier to just skip the documentation and go straight to the [examples](examples/README.md) as soon as possible. :wink:

- [Initializing the Service](#initializing-the-service)
- [Configuring the Input](#configuring-the-input)
- [Configuring the Commands](#configuring-the-input)
    - [Adding Commands](#adding-commands)
    - [Adding Foreach](#adding-foreach)
    - [Adding HTTP Requests](#adding-http-requests)
    - [Adding HTTP Foreach](#adding-http-foreach)
- [Configuring the Output](#configuring-the-output)
- [Compiling the Service](#compiling-the-service)
- [Advanced Features](#advanced-features) 
    - Delete method
    - using output from former command
    - Direktiv file
- [Custom Go Code](#custom-go-code)

## Initializing the Service

The first step starting with a new service is initializing the project. Direktiv's service builder comes with a docker container so no local installation is required. To initialise a project simply call the following command:

```
docker run -v `pwd`:/tmp/app direktiv/service-builder init myservice
```

> **TIP**: On Linux the owner of the created files is `root`. To avoid this pass in a user and group.

*Passing user and group id to the container*

```
docker run --user 1000:1000 -v `pwd`:/tmp/app direktiv/service-builder init myservice
```

The container maps a local directory to the container and uses it as the base. In the above example, we used the Linux `pwd` command, but this can also be a static file path. The service name is the last argument. There will be a new folder named v1.0.0 in the designated target folder after launching the container, which contains four files.

**1. Dockerfile**

A multi-stage build is used in the default Dockerfile. The generated application is built in the first stage. The second stage may be customized to meet the requirements of the service, for example through adding additional files to the base image or change the base image via the `FROM` directive. The last three lines should not be altered.

```Dockerfile
FROM golang:1.18-buster as build

COPY go.mod src/go.mod
COPY go.sum src/go.sum
RUN cd src/ && go mod download

COPY cmd src/cmd/
COPY models src/models/
COPY restapi src/restapi/

RUN cd src && \
    export CGO_LDFLAGS="-static -w -s" && \
    go build -tags osusergo,netgo -o /application cmd/myservice-server/main.go; 

FROM ubuntu:21.04

RUN apt-get update && apt-get install ca-certificates -y

# DONT CHANGE AFTER THIS 
COPY --from=build /application /bin/application

EXPOSE 8080

CMD ["/bin/application", "--port=8080", "--host=0.0.0.0"]
```

**2. swagger.yaml**

This file is the configuration file for the service and the configuration options will be explained in this documentation. An example can be seen [here](examples/bash/v1.0.0/swagger.yaml).

**3. run.sh**

This helper script builds the service and container and starts it. This script can be used for faster development and testing. Can be used after the service has been compiled the first time. 

**4. go.mod**

Manages the go dependencies. Should not be altered. 

## Configuring the Input

The input values required for a new service must be planned. Although the service builder may be set to accept all JSON data, using an input configuration prevents the service from malfunctioning and delivering incorrect results.

In the `swagger.yaml` file generated at the beginning, the default input configuration begins at `paths:`. This part of the document follows the [swagger specification](https://swagger.io/docs/specification/describing-parameters/) exactly, and it contains no Direktiv-specific extensions. Nevertheless Direktiv only uses the base path `/` with POST for services so only the body section is important here and changes should be made here only. For example the default `swagger.yaml` has on mandatory `name` parameter of type `string` configured in the body.

```yaml
...
- name: body
  in: body
  schema:
    type: object
    required:
      - name
    properties:
      name:
        type: string
        example: YourName
        description: The full name for the greeting
...
```

If the service requires e.g. a list of integers it could be changed to the following which make the a list of integers optional.

```yaml
...
- name: body
  in: body
  schema:
    type: object
    additionalProperties: false
    required:
      - name
    properties:
      name:
        type: string
        example: YourName
        description: The full name for the greeting
      values: 
        type: array
        items: 
        type: integer
...
```

With this configuration additional properties sent to the service from Direktiv will be ignored and can not be used in templating later in the command section. This behaviour can be changed with the attribute `additionalProperties`. If set to `true` the data will be available. Alternatively the input can be unspecified without any checks. This would accept every JSON payload from Direktiv.

> **TIP**: If the `additionalProperties` approach is used the variables of the incoming request can not be access directly e.g. `{{ .Name }}` but in an wrapper variable `In` because the data becomes a map. It has to be access with `{{ index .In "name" }}`.

```yaml
- name: body
  in: body
  schema:
    type: object
    additionalProperties: {}
```

To make a logical link between this input configuration and the Direktiv service, the following three examples provide an input definition for a service and how it converts to a Direktiv workflow. The last example demonstrates how this looks as a request made by Direktiv to the service.

*1. Input Definition*
```yaml
- name: body
  in: body
  schema:
    type: object
    properties:
      names: 
        type: array
        items: 
          type: string
      values:
        type: integer
```

*2. Direktiv Action State*
```yaml
- id: run
  type: action
  action:
    function: myservice
    input:
      name: Diana
      values:
      - 100
      - 200
      - 300
```

*3. Service Payload*
```json
{
  "name": "jens",
	"values": [
		1,
		2,
		3
	]
}
```

## Configuring the Commands

The commands to be executed are defined in the `x-direktiv` section under `cmds`. Multiple commands can be configured and the response will be an array of the results of each command.

*Example Configuration in swagger.yaml*
```yaml
x-direktiv:  
  cmds:
  - action: exec
    exec: echo 'Hello'
  - action: exec
    exec: echo 'World'
```

*Example Service Response*
```json
[
	{
		"result": "Hello",
		"success": true
	},
	{
		"result": "World",
		"success": true
	}
]
```

In almost all attributes in this section go templating can be used. During runtime the values provided in the attributes, e.g. `exec` are getting parsed with the data configured in the [input](#configuring-the-input) section. If a e.g. a `name` attribute is configured in the input section it can be used as `{{ .Name }}`. 

The nature of the template engine requires the attributes to be upper-case. If there are problems during development there is a `debug` field which can be used to add additional debug information to the application. In particular the input and input data for the templating are printed to the console. 

**Enabling debug**
```yaml
 x-direktiv:  
  # enable debug output
  debug: true 
  cmds:
  - action: exec
    exec: echo Hello
```

The Service Builder supports all go templating commands plus the extension library [sprig](http://masterminds.github.io/sprig/). That means within the fields even `if` or `range` statements can be executed (The `-` in `{{- }}` avoids new lines). 

**Example 'if' Statement**
```yaml
x-direktiv:  
  cmds:
  - action: exec
    exec: |- 
      echo {{- if eq (deref .Name) "Mike" }} Go away Mike{{- else }} Hello {{ .Name }}{{- end }}
```

> **Note:** For string or integer comparison the values need to be dereferenced with `deref`. The `debug` can provide more details if there are any problems with templating.

**Example 'range' Statement**
```yaml
x-direktiv:  
  cmds:
  - action: exec
    exec: echo  Hello {{- range $i,$n := .Names }} {{ $n }} {{- end }}
```

By default these commands are getting returned as array of commands. There is an additional `output` field which will be used as service response if it is configured. This is usefule if multiple commands or scripts need to run but the response should not show all the executed commands. This seciton has to be configured as JSON response but templating can be used here as well. 

**Example Output**
```yaml
x-direktiv:  
  cmds:
  - action: exec
    exec: echo Hello {{ .Name }}
  output: |
    {
      "this-is-the-hello-value": "{{ index (index . 0) "result" }}"
    }
```

The `output` instruction has access to all executed commands. As seen before all results are in array. The previous exmplae accesses the first command result with `(index . 0)`. The command result is a map with the values `result` and `success` so to fetch the result it uses `index` again. In case the result is JSON and not text the templating can access indivdiual values/attributes from the JSON. As with the other fields `if` statements and other template instructions can be used as well.

```yaml
x-direktiv:  
  cmds:
  - action: exec
    exec: echo Hello {{ .Name }}
  output: |
    { 
      {{- if index (index . 0) "success" }}
            "this-is-the-hello-value": "{{ index (index . 0) "result" }}"
      {{- else }}
            "did": "not work"
      {{- end }}
    }
```

## Adding Commands

There are four different types of commands implemented: `exec`, `foreach`, `http` and `httpForeach`. The next section will explain those commands and their configuration. 

The `exec` command executes a command with command line arguments and other configuration options. It is the most powerful of the four different types. The following example uses all available configuration options:

```yaml
cmds:
- action: exec
  exec: bash -c 'echo { \"hello\": \"{{ .Hello }}\"  > /tmp/myfile }; env'
  error: my.small.error
  output: /tmp/myfile
  continue: true
  print: true
  silent: {{ .DoPrint }}
  env: ["KEY=value", "NAME={{ .Name }}"]
```

The folowing attributes can be used:

### error

If a command fails, Direktiv informs the user with an error code. The default value of `io.direktiv.command.error` is used for reporting problems if this field isn't defined. Otherwise this is used as error code. This is the only place where templating cannot be utilized.
  
### output

This field reads the nominated file and uses this as reposnse for the command. The default behaviour is to use the output from standard out but this field changes that behaviour. In the above example the `exec` instruction writes the result in `/tmp/myfile` and this will be used as result for this command. This is not the overall response for the service, just for this command.

### continue

This field controls the behaviour in case of an error. The default behaviour is to throw an error if the execution fails. If continue is set to `true` the execution is marked as failed but the service continues with either executing the next command or returning if it is the last or only one.  

### print

By default the command with all arguments is getting printed to the log file. In cases where sensitive data is part of the command it can be supressed with setting `print` to false.

### silent

A command can be executed in `silent` mode which means it is not printing to the logs. The command supresses all output.

### env

This defines the environment variables for the command. Templating can be used for the keys as well as the values.


## Adding Foreach

In cases where the same command needs to be executed over a list of items the `foreach` execution type can be used. The configuration is almost identical to the `exec` execution. There is an additional attribute `loop` to define the array to use for iteration. This array needs to be defined in the [input](#configuring-the-input) section.

```yaml
x-direktiv:  
  cmds:
  - action: foreach
    loop: .Names
    exec: echo Hello {{ .Item }}
```

The result is an array of results and will look similar to the following JSON snippet.


```json
[
	[
		{
			"result": "Hello Mike",
			"success": true
		},
		{
			"result": "Hello Steven",
			"success": true
		},
		{
			"result": "Hello Sarah",
			"success": true
		}
	]
]
```

## Adding HTTP Requests

Although a HTTP request could be achieved with a `curl` command it is provided as a convenient command. It simply executes a HTTP request and returns the reponse including the headers. The configuration is simple and the following example shows all the configuration options.

```yaml
- action: http
  url: http://www.direktiv.io/{{ .Path }}
  method: get
  headers: 
    - Content-Type: application/json
  username: hello
  password: world
  insecure: true 
  errorNo200: true
  continue: true
  data: |
    {
      "myname": "{{ .Name }}"
    }
```

The folowing attributes can be used:

### url / method

The URL and method to use for this request. This defines the actual request.

### username / password

If both values are set they are used for Basic Authentication. 

### insecure

If the backend uses self-signed certificates this setting ignores any SSL errors. Not recommended for production use. 

### errorNo200

If there is a reponse from the taregt URL by default it counts as success. For example a 404 is a successful request because technically the request got a response. If this attribute is set to true every status code above 299 is considered as error.

### continue

Continue has the same behaviour as in the `exec` context. If the request fails the next command will stil be executed if this is set to `true`.

### data

The data section provides the body if the request is a POST request. The `data` atttribute is for plain text content like JSON or XML. Templating can be used to dynamically change the content based on theinput of the service. 

### data64

For binary data this attribute is available. If set the data will be decoded and sent as body of the request. The base64 can be data coming from the input or a special template function `file64` reads a file from the filesystem and attaches it. 

```yaml
- action: http
  url: http://www.direktiv.io/{{ .Path }}
  method: get
  data64: {{ file64 myfile.tar.gz }}
```
