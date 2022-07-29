 # Direktiv Function Builder

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
    - [Delete Method](#delete-method) The swaggger UI is available 
    - [Chaining Commands](#chaining-commands) 
    - [Direktiv File](#direktiv-file) 
- [Secrets](#secrets) 
- [Generate Documentation](#generate-documentation)
- [Testing](#testing)

## Quickstart

The first step to create a function with the function builder is to download the latest release from the [release page](https://github.com/direktiv/function-builder/releases). It is available for Linux, Windows and Mac. 

The second step is to create an empty directory or cloning an empty Git project. All the function builder commands have to be executed in this directory. There are basically only three commands to build a function.

The first one is only called once to prepare the project folder. It creates the required files to generate a function later. The most important files are the `Dockerfile` and the `swagger.yaml`.

*Preparing a function*
```
function-builder prepare -f my-function
```

The whole process of building a function with this tool is to change the `swagger.yaml` file and run the second command which generates the source code based on the information in the `swagger.yaml` file. The default template will create a working example function.

*Generating function source code*
```
function-builder gen
```

After these two steps the `run.sh` script can build the function and serve it for local testing. The function is available at [http://localhost:9191](http://localhost:9191) and the swagger UI at [http://localhost:9191/docs](http://localhost:9191/docs). The function can be tested via `curl` as well.

*Curl the function*
```
curl -X POST -H "Direktiv-ActionID: development" http://localhost:9191/
```

This request should return something like `{"my-function":null}`. The default template takes a list of commands as parameter and because no command was provided it return `null`. If a command is provided it returns the output of that command.

*Curl the functio with command*
```
curl -X POST -H "Direktiv-ActionID: development" \
  -H "Content-Type: application/json" \
  -d '{ "commands": [ { "command": "echo Hello" } ] }' http://localhost:9191/
```

## Initializing the Service

After running the prepare step with `function-builder prepare -f my-function` there are multiple files in the function directory. This step can only be called once at the start of building the function. This section lists the create files and what how to use them. 

**1. Dockerfile**

A multi-stage build is used in the default Dockerfile. The generated application is built in the first stage. The second stage may be customized to meet the requirements of the service, for example through adding additional files to the base image or change the base image via the `FROM` directive. The last three lines should not be altered.

```Dockerfile
FROM golang:1.18.2-alpine as build

WORKDIR /src

COPY build/app/go.mod go.mod
COPY build/app/go.sum go.sum

RUN go mod download

COPY build/app/cmd cmd/
COPY build/app/models models/
COPY build/app/restapi restapi/

ENV CGO_LDFLAGS "-static -w -s"

RUN go build -tags osusergo,netgo -o /application cmd/kubectl-server/main.go; 

FROM ubuntu:22.04

RUN apt-get update && apt-get install ca-certificates -y

# DON'T CHANGE BELOW 
COPY --from=build /application /bin/application

EXPOSE 8080

CMD ["/bin/application", "--port=8080", "--host=0.0.0.0", "--write-timeout=0"]
```

**2. swagger.yaml**

This file is the configuration file for the function and the configuration options will be explained in this documentation. An example can be seen [here](examples/bash/v1.0.0/swagger.yaml).

**3. run.sh**

This helper script builds the function and container and starts it. This script can be used for faster development and testing. Can be used after the function has been compiled the first time. It starts it at http://localhost:9191

**4. LICENSE**

Default Apache 2.0 license for the function.

**5. .gitignore**

Default .gitignore file with entries for the template directory and `.direktiv.yaml`.

## Configuring the Input

The input values required for a new function must be planned. Although the function builder may be set to accept all JSON data, using an input configuration prevents the function from malfunctioning and delivering incorrect results.

In the `swagger.yaml` file generated at the beginning, the default input configuration begins at `paths:`. This part of the document follows the [swagger specification](https://swagger.io/docs/specification/describing-parameters/) exactly, and it contains no Direktiv-specific extensions. Nevertheless Direktiv only uses the base path `/` with POST for function so only the body section is important here and changes should be made here only. For example the default `swagger.yaml` has on mandatory `commands` parameter of type `array` configured in the body. This can be changed to whatever is required for the function.

```yaml
...
commands:
  type: array
  description: Array of commands.
  items:
    type: object
    properties:
      command:
        type: string
        description: Command to run
...
```

If the function requires e.g. a list of integers it could be changed to the following which make the a list of integers mandatory.

```yaml
...
- name: body
  in: body
  schema:
    type: object
    additionalProperties: false
    required:
      - values
    properties:
      values: 
        type: array
        items: 
        type: integer
...
```

With this configuration additional properties sent to the function from Direktiv will be ignored and can not be used in templating later in the command section. This behaviour can be changed with the attribute `additionalProperties`. If set to `true` the data will be available. Alternatively the input can be unspecified without any checks. This would accept every JSON payload from Direktiv.

> **TIP**: If the `additionalProperties` approach is used the variables of the incoming request can not be access directly e.g. `{{ .Name }}` but in an wrapper variable `In` because the data becomes a map. It has to be access with `{{ index .In "name" }}`.

```yaml
- name: body
  in: body
  schema:
    type: object
    additionalProperties: {}
```

To make a logical link between this input configuration and the Direktiv function, the following three examples provide an input definition for a function and how it converts to a Direktiv workflow. The last example demonstrates how this looks as a request made by Direktiv to the function.

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
    function: myfunc
    input:
      name: Diana
      values:
      - 100
      - 200
      - 300
```

*3. Function Payload*
```json
{
  "name": "Diana",
	"values": [
		100,
		200,
		300
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

*Example Function Response*
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

The function builder supports all go templating commands plus the extension library [sprig](http://masterminds.github.io/sprig/). That means within the fields even `if` or `range` statements can be executed (The `-` in `{{- }}` avoids new lines). 

**Example 'if' Statement**
```yaml
x-direktiv:  
  cmds:
  - action: exec
    exec: |- 
      echo {{- if eq (deref .Name) "Mike" }} Go away Mike {{- else }} Hello {{ .Name }}{{- end }}
```

> **Note:** For string or integer comparison the values need to be dereferenced with `deref`. The `debug` can provide more details if there are any problems with templating.

**Example 'range' Statement**
```yaml
x-direktiv:  
  cmds:
  - action: exec
    exec: echo  Hello {{- range $i,$n := .Names }} {{ $n }} {{- end }}
```

By default these commands are getting returned as array of commands. There is an additional `output` field which will be used as function response if it is configured. This is useful if multiple commands or scripts need to run but the response should supress some executed commands. This section has to be configured as JSON response but templating can be used here as well. 

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

The `output` instruction has access to all executed commands. As seen before all results are in array. The previous example accesses the first command result with `(index . 0)`. The command result is a map with the values `result` and `success`. To fetch the result it uses `index` again. In case the result is JSON and not text the templating can access indivdiual values/attributes from the JSON. As with the other fields `if` statements and other template instructions can be used as well.

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

This field controls the behaviour in case of an error. The default behaviour is to throw an error if the execution fails. If continue is set to `true` the execution is marked as failed but the service continues with either executing the next command or returning if it is the last or the only one.  

### print

By default the command with all arguments is getting printed to the log file. In cases where sensitive data is part of the command it can be supressed with setting `print` to false.

### silent

A command can be executed in `silent` mode which means it is not printing to the logs. The command supresses all output.

### env

This defines the environment variables for the command. Templating can be used for the keys as well as the values.

### runtime-envs

If runtime environment variables are required, meaning the client sends environment variables to the service, the `runtime-envs` attribute can be used to add them to the static list in `env`. This attibute contains a template which has to return a JSON array of strings in `KEY=VALUE` format. The folowing example is taken from the terraform service. 

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

The default behaviour of functions is to return all JSON data generated by the commands. Sometimes it is required to guarantee a certain response. 

**Default Response**
```yaml
responses:
  200:
    schema:
      type: object
      additionalProperties: 
```

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

## Compiling and runnig the function

After every change in the `swagger.yaml` file configuration the function needs to be generated. This can be done with a simple command and the command must be run where the `prepare` command was executed. 

```
function-builder gen
```

The `prepare` phase created a convenience file `run.sh` which can be used to run and test the function. It compiles the function and starts a container exposing port `9191`. After starting it the function can be used with a simple `curl` or via the auto-generated docs/ui at [http://127.0.0.1:9191/docs](http://127.0.0.1:9191/docs). 

**Runnning curl**
```
curl -X POST -H "Direktiv-ActionID: development" -H "Content-Type: application/json"  http://127.0.0.1:9191 -d '{ "name": "myname" }'
```

## Advanced Features 

### Delete Method

If a service is getting cancelled in Direktiv it sends a DELETE request with the action ID to the cancelled service. The Direktiv Service Builder will stop the executing command if that happens. If there is a requirement to run additional commands after it can be configured in `swagger.yaml`. The `Direktiv-ActionID` header contains the action id of the instance.


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

## Secrets

To define required secrest in the function there is an additional section in the `swagger.yaml` file called `x-direktiv-secrets`. This section is not required but is important for [testing](#testing) and documentation. It is a simple list of the secrets and its description.

```yaml
x-direktiv-secrets:
  - name: myPassword
    description: Password to open the chest
```


## Generate Documentation

Markdown documentation can be generated based on the `swagger.yaml` file. To get a complete documentation is recommended to provide examples and descriptions to input and response parameters. The [http-request](https://github.com/direktiv-apps/http-request/blob/main/swagger.yaml) is an example of a complete `swagger.yaml` file for documentation purposes. 

There are new Direktiv specific attributes to improve the auto-generated documentation. These have no impact on the functionality of the service and are for documentation only:

#### x-direktiv-function

This attribute can be used to provide a copy-and-paste example of the function to use in Direktiv workflows. It will be used to generate tests as well.

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
function-builder docs
```

## Testing

During the generation phase function tests are getting generated based on the information provided. There are five files in the `test` directory. These files are not geting overwritten. If e.g. secrets change the files need to be manually deleted.

**karate.yaml.test.feature**

This file is primarily for local testing. It use the [Karate API testing](https://github.com/karatelabs/karate) framework. If a service has been started with the `run.sh` file there is an additional `run-tests.sh` file which runs the tests in this file against the running function instance. The default behaviour is to run all tests in that file but individual scenarios can be tested adding `--name=scenario-name` to `run-tests.sh`. 

**karate.yaml**

To run thet test in Direktiv directly this karate.yaml flow is getting created. The whole project can be added as Direktiv namespace and tested within Direktiv itself. This flow requires a `host` parameter to define the target to test against. If it is locally that can be the local IP of the computer running the function with `run.sh`. The reason for this file is to include it in e.g. CI/CD flows. 

**tests.yaml**

This is a flow using the function in a regular flow in Direktiv. By default the image used has the pattern `gcr.io/direktiv/apps/<MYFUNCTION>:test`. This can be changed to e.g. `localhost:5000/<MYFUNCTION>` for local tests. 

**karate-event.yaml / tests-event.yaml**

This are support flows if the tests are getting triggered by events within Direktiv. It is useful if multiple functions are part of CI/CD and the tests are getting triggered by events with the function name. 




