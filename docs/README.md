 # Direktiv Service Builder

Creating a new app takes three steps: configuring the input, the command, and output. This documentation will go through these three stages in detail, as well as all configuration choices. As a good developer, you may believe it's easier to just skip the documentation and go straight to the [examples](examples/README.md) as soon as possible :wink: or look at the [direktiv-apps repository](https://github.com/direktiv-apps). 

- [Quickstart](#Quickstart)
- [Initializing the Service](#initializing-the-service)
- [Configuring the Input](#configuring-the-input)
- [Configuring the Commands](#configuring-the-input)
    - [Adding Commands](#adding-commands)
    - [Adding Foreach](#adding-foreach)
    - [Adding HTTP Requests](#adding-http-requests)
    - [Adding HTTP Foreach](#adding-http-foreach)
- [Configuring the Output](#configuring-the-output)
- [Compiling and Running the Service](#compiling-and-running-the-service)
- [Advanced Features](#advanced-features) 
    - [Delete Method](#delete-method) 
    - [Chaining Commands](#chaining-commands) 
    - [Direktiv File](#direktiv-file) 
- [Custom Go Code](#custom-go-code)
- [Generate Documentation](#generate-documentation)

## Quickstart

The initial `prepare` and `gen` steps are already creating a working demo service. To test the service builder functionality download the service builder from the Github release page and run the following commands:

```
# creates dockerfile, run.sh, swagger.yaml, go.mod
docker run --user 1000:1000 -v `pwd`:/tmp/app direktiv/service-builder init myservice

# generates the source code
docker run --user 1000:1000 -v `pwd`:/tmp/app direktiv/service-builder gen

# compiles the source code and builds container
./run.sh
```

This will init, compile, build and start the service. It is accessible on [127.0.0.1:8080](127.0.0.1:8080):

```
curl -X POST -H "Direktiv-ActionID: development" -H "Content-Type: application/json"  http://127.0.0.1:8080 -d '{ "name": "myname" }'
```

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

The container maps a local directory to the container and uses it as the base. In the above example, we used the Linux `pwd` command, but this can also be a static file path. The service name is the last argument. In the designated target folder, after running the container, will be four files:

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

In almost all attributes in this section go templating can be used. During runtime the values provided in the attributes, e.g. `exec` are getting parsed with the data configured in the [input](#configuring-the-input) section. If a e.g. a `name` attribute is configured in the input section it can be used as `{{ .Name }}`. The variable `DirektivDir` is always set and containes the working directory of the function.

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

### runtime-envs

If runtime environment variables are required, meaning the client sends environment variables to the service, the `runtime-envs` attribute can be used to add them to the static list in `env`. Ths attibute contains a template which has to return a JSON array of strings in `KEY=VALUE` format. The folowing example is taken from the terraform service. 

```yaml
runtime-envs: |
  [
  {{- range $index, $element := .Body.Args }}
  {{- if $index}},{{- end}}
  "TF_VAR_{{ $element.Name }}={{ $element.Value }}"
  {{- end }}
  ]
```

## Adding Foreach

In cases where the same command needs to be executed over a list of items the `foreach` execution type can be used. The configuration is almost identical to the `exec` execution. There is an additional attribute `loop` to define the array to use for iteration. This array needs to be defined in the [input](#configuring-the-input) section.

In foreach commands the object used in templating provides two values which can be used. One is `.Item` which is the actual object in the loop. The whole request object is in `.Body` if the templating values need access to values outside of the actual loop value. 

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

### headers / runtime-headers

Static headers can be set as a key/value pair where the value can be still use templating. For additional runtime headers the `runtime-headers` attribute can be used. This has to be a template value pointing to a key/value array in the payload.

```yaml
runtime-headers: .AdditionalHeaders
```

### username / password

If both values are set they are used for Basic Authentication. 

### insecure

If the backend uses self-signed certificates this setting ignores any SSL errors. Not recommended for production use. 

### errorNo200

If there is a reponse from the taregt URL by default it counts as success. For example a 404 is a successful request because technically the request got a response. If this attribute is set to true every status code above 299 is considered as error.

### continue

Continue has the same behaviour as in the `exec` context. If the request fails the next command will stil be executed if this is set to `true`.

### data

This can be used to add a payload to a POST request. This parameter is an object with two attributes `kind` and `value`. The first value defines what type of data is being posted.

- string: Plain string data. Templates can be used with this kind. Value is the string to use.
- base64: Base64 content will be converted into a byte array and send with the request. Value is the Base64 string. 
- file: A file will be attached. Value is the file name. 

```yaml
x-direktiv:  
  debug: true
  cmds:
  - action: http
    url: https://myurl.com
    method: post
    # base64 example
    data:
      kind: base64
      value: '{{ .Datain }}'
    # string example
    data:
      kind: string
      value: 'Send this: {{ .Datain }}'
    # file example
    data:
      kind: file
      value: direktiv-file.txt
```

## Adding HTTP Foreach

HTTP Foreach uses the same attributes as the single HTTP request. It needs a `loop` attribute to define tha array to iterate over. All templating rules are the same as they are for `foreach`. The actual item value of the iteration is in `.Item` and the input data overall is in `.Body`.

```yaml
- action: foreachHttp
  loop: .Names
  url: http://www.direktiv.io/{{ .Item }}
  method: get
  headers: 
    - Content-Type: application/json
  data: |
    {
      "myname": "{{ .Body.Name }}"
    }
```

## Configuring the Output

The default behaviour of Direktiv Servbice Build is to return all JSON data generated by the commands. Sometimes it is required to guarantee a certain response. 

**Default Response**
```yaml
responses:
  200:
    schema:
      type: object
      additionalProperties: {}
```

The above example accepts all JSON properties. If it is required to validate the output it can be modified the same way as the [input](#configuring-the-input). The following example shows a response which requires `my-return-value` as return value. Because it is required the service would fail if that value does not exist. If the `required` is removed, the service would return an empty JSON object if the value does not exist. 

**Validate Response**
```yaml
responses:
  200:
    schema:
      type: object
      required:
        - my-return-value
      properties: 
        my-return-value: 
          type: string
```

## Compiling and Running the Service

After every change in the `swagger.yaml` file configuration the service needs to be generated. This can be done with a simple command and the command must be run where the `init` command was executed. 

```
docker run --user 1000:1000 -v `pwd`:/tmp/app direktiv/service-builder gen
```

The `init` phase created a convenience file `run.sh` which can be used to run and test the service. It compiles the service and starts a container exposing port `8080`. After starting it the service can be used with a simple `curl` or via the auto-generated docs/ui at [http://127.0.0.1:8080/docs](http://127.0.0.1:8080/docs). 

**Runnning curl**
```
curl -X POST -H "Direktiv-ActionID: development" -H "Content-Type: application/json"  http://127.0.0.1:8080 -d '{ "name": "myname" }'
```

DIREKTIV


## Advanced Features 

### Delete Method

If a service is getting cancelled in Direktiv it sends a DELETE request with the action ID to the cancelled service. The Direktiv Service Builder will stop the executying command if that happens. If there is a requirement to run additional commands after it can be configured in `swagger.yaml`. The `Direktiv-ActionID` header contains the action id of the instance.


**Delete Configuration**
```yaml
paths:
  /: 
    delete:
      parameters:
        - name: Direktiv-ActionID
          in: header
          type: string
          description: |
            On cancel Direktiv sends a DELETE request to
            the action with id in the header
      x-direktiv:
        cancel: echo 'cancel {{ .DirektivActionID }}'
```

### Chaining Commands

It is possible to use output of a previous command if multiple commands are used. The commands are stored in an array `Commands` and can be accessed with their index. 


**Using Command Results**
```yaml
cmds:
- action: exec
  exec: echo 'Hello {{ .Name }}'
- action: exec
  exec: echo 'The greeting was {{ index (index .Commands 0) "result" }}'
```

## Custom Go Code

In cases where something very specific needs to be build the Service Builder can generate go skeleton code. The service needs to run the `init` command as well but the generate command is slightly different and `gen-custom` is used instead of `gen`.

```
docker run --user 1000:1000 -v `pwd`:/tmp/app direktiv/service-builder gen-custom
```

This generates all files similar to the code generation except two files:

- restapi/operations/direktiv-post.go
- restapi/operations/direktiv-delete.go

They will not be overwritten and the implementation of the service need to be implemented in those files. The `direktiv-post.go` includes two functions. `PostDirektivHandle` is the actual implementation of the service whereas `HandleShutdown` is a function called before the Direktiv service gets scheduled out by Knative. 

The other file uses the function `DeleteDirektivHandle` to handle cancel requests from Direktiv. This function does not need to be implemented. Only if there is a requirement to handle cancel commands.

## Generate Documentation

Markdown documentation can be generated based on the `swagger.yaml` file. To get a complete documentation is recommended to provide examples and descriptions to input and response parameters. The [http-request](https://github.com/direktiv-apps/http-request/blob/main/swagger.yaml) is an example of a complete `swagger.yaml` file for documentation purposes. 

There are new Direktiv specific attributes to improve the auto-generated documentation. These have no impact on the functionality of the service and are for documentation only:

#### x-direktiv-function

This attribute can be used to provide a copy-and-paste example of the function to use in Direktiv workflows.

```yaml
...
paths:
  /: 
    post:
      x-direktiv-function: |-
        functions:
          - id: request
            image: direktiv/http-request
            type: knative-workflow
      parameters:
...
```

#### x-direktiv-examples

Examples can be helpful and in this section full examples of the usage can be provided to users of the service. They are in YAML but stored as string so comments can be added as well if required. 

```yaml
...
paths:
  /: 
    post:
      x-direktiv-examples:
        - title: Basic
          content: |-
            - id: req
                 type: action
                 action:
                   function: request
                 input: 
                 url: "http://www.direktiv.io"
        - title: Post Request
          content: |-
            url: http://www.direktiv.io
            method: post
      parameters:
...
```


#### x-direktiv-errors

In this section custom errors can be added if they have been defined in one of the commands or custom go code has been written for the service. This

```yaml
...
paths:
  /: 
    post:
      parameters:
      ...
      x-direktiv-errors:
        io.direktiv.command.error: Command execution failed
        io.direktiv.output.error: Template error for output generation of the service
        io.direktiv.ri.error: Can not create information object from request
...
```

The doumentation can be generated with one command execute in a directory with a `swagger.yaml` file. This generates a README.md file for the service. 

```
docker run --user 1000:1000 -v `pwd`:/tmp/app direktiv/service-builder docs
```