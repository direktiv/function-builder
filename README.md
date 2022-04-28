# Direktiv Action Builder

During a flow execution, [Direktiv](https://github.com/direktiv/direktiv) is utilizing containers to execute actions. Although there are many actions already available sometimes it is required to write custom actions. 

Custom functions are often built by combining other pieces of code together, and this action builder tool can assist you to generate the necessary source code for them. In most situations, no development is required, but it is possible to supply custom code to the builder and have it generate the needed wrapper functions.

The action builder is built on OpenAPI, which serves as a standard to generate and execute the actions.

## Quickstart, tl;dr

- Initialise the application which generates the basic configuration.

```
docker run -v `pwd`:/tmp/app direktiv/action-builder init myapp
```

- Tweak the configuration to do what you need it to do.

- Generate the source code

```
docker run -v `pwd`:/tmp/app direktiv/action-builder gen v1.0.0
```

- Test the application

Execute `run.sh` within the v1.0.0 folder which was generated during the `init` step. The following command starts the server on port 8080 and the docs are available under [http://127.0.0.1:8080/docs](http://127.0.0.1:8080/docs).

```
cd v1.0.0 && ./run.sh
```

The service call can be tested with a simple curl command:

```
curl -X POST -H "Direktiv-ActionID: development" -H "Content-Type: application/json"  http://127.0.0.1:8080 -d '{ "name": "myname" }'
```

- Publish the application

The last step is to push the application to a container registry and use it within a Direktiv workflow.

```
docker build -t myname/myapp . && docker push myname/myapp
```

The action can be used in a flow like this:

```yaml
functions:
- id: myapp
  image: myname/myapp
  type: knative-worklflow
states:
- id: myapp 
  type: action
  action:
    function: myapp
    input: 
      name: myname
```

## Documentation

The action build process is a two step process. The first step is to initialise the application and the second step is to generate the source code. 

### Init an action

The first step is to initialize the action with the init command. It is important to map the folder where the action should reside into the app builder container. The directory in the container is `/tmp/app`.

```
docker run -v `pwd`:/tmp/app direktiv/action-builder init myapp
```

This command creates a folder with the default version number `1.0.0` in the target directory. Within that directory is a simple hello-world application with the following files:

- Dockerfile

The Dockerfile contains a two step build process. The first step compiles the action. The second part is the actual container which will run in Direktiv. The `FROM` statement can be changed and additional docker commands and files can be added. The final `CMD` instruction should *NOT* be changed.

- go.mod

Basic go.mod file. No change required. 

- run.sh

This shell script is provided for easy testing. It compiles and starts the container so it can be tested outside of Direktiv. 

- swagger.yaml

This is the file which is the base for the generated service and will be explained in detail in the next section. 

**Docker File Permissions**: Docker runs as root inside the container. It can be required to run the command with a different user to avoid permission issues, e.g.:

```
docker run --user 1000:1000 -v `pwd`:/tmp/app direktiv/action-builder init myapp
```

### Configuring the action

The action builder is using the swagger file to generate the source code. There are two methods defined. The `post` method is used to execute the action and the `delete` method should be used to cancel a running action. 

The parameter section defines the input for the action. The headers `Direktiv-ActionID` and `Direktiv-TempDir` are provided during a call from a Direktiv flow. The `body` section is sent to the action to be consumed. Please read the [swagger documentation](https://swagger.io/docs/specification/2-0/describing-request-body/) for further details how to configure input parameters in the body of a request. The default action requires a `name` attribute. 

```yaml
    post:
      parameters:
        - name: Direktiv-ActionID
          in: header
          type: string
          description: |
            direktiv action id is an UUID. 
            For development it can be set to 'development'
        - name: Direktiv-TempDir
          in: header
          type: string
          description: |
            direktiv temp dir is the working directory for that request
            For development it can be set to e.g. '/tmp'
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
```

If attributes are not defined in the input section they are not available in templating later. If an unspecified input object is required `additionalProperties` can be used and the input object will be a map of values.

```yaml
- name: body
  in: body
  schema:
    type: object
    additionalProperties: {}
```

There is one special attribute for the input parameters and this has the type `DirektivFile`. Direktiv can provide files to actions based on [variables](https://docs.direktiv.io/v0.6.1/getting_started/persistent-data/#demo) but sometimes the actions need a file, e.g. a small shell script or a token where a variable is not really needed. This type reads the string input and creates a file with the name provided.

```yaml
- name: body
  in: body
  schema:
    type: object
    required:
      - name
    properties:
      script:
        $ref: "#/definitions/direktivFile"
```

The payload for this file would be like this:

```json
{
	"script": {
		"data": "data in file",
		"name": "file.txt"
	}
}
```

The next `x-direktiv` section in the post method defines what the action will execute. The `cmds` section can have one or more commands. The action builder will execute them in the order they are configured. *The commands, e.g. echo, need to be available on the image used in the Dockerfile of the action*.

```yaml
x-direktiv:  
  cmds:
  - action: exec
    exec: echo 'Hello {{ .Name }}'
  - action: exec
    exec: sleep 10
```

This example would execute two commands, echo and sleep. The commands can use variables via [go templates](https://pkg.go.dev/text/template) and [sprig](http://masterminds.github.io/sprig/). The variables are based on the input parameters for the action. In the default action the input parameter is `name` so it can be used within the commands. *The parameters can be defined lower-case but have to be used title-case, e.g. input `name` is `Name` in templates*. All go template commands are supported so even if/else statements can be used. A debug mode is available to see the templates and data for the data parsing in case the action is not returning the expected result (`run.sh` is recommended for debugging).

```yaml
x-direktiv:
  debug: true
  in: body
  schema:
    type: object
    properties:
      names:
        type: array
        items:
          type: string
``` 

The default response is a list of executed commands named `cmdX`. The reponse for the example would be the following JSON payload.


```json
{
  "return": {
    "cmd0": {
      "result": "Hello jens",
      "success": true
    },
    "cmd1": {
      "result": "",
      "success": true
    }
  }
}
```

The output of the action can be modified with a template if the default response does not statisfy the requirements. The output can be configured in the `output` section. The parameter for the template is the response object, which is a map of the commands execute with the keys `cmdX` for each command.

```        
output: |
  {
    "greeting": "{{ index (index . "cmd0") "result" }}"
  }
```

The last part of the definition of an action is the response. By default it returns JSON without testing the type with the following definition.

```yaml
responses:
  200:
    description: nice greeting
    schema:
      type: object
      additionalProperties: {}
```

To guarantee a response type the response object can be defined as well. If the generated JSON from the commands don't match the response will be empty. 

```yaml
responses:
  200:
    description: nice greeting
    schema:
      type: object
      properties:
        greeting:
          type: string
```

In the `cmds` of the `x-direktiv` one or multiple commands can be defined. These commands can have a different `action` type. Each type has different additional attributes to change the behaviour. The following four types can be used.

### Execute 

The action `exec` runs a command with the arguments defined in `exec`. 

```yaml
x-direktiv:  
  cmds:
  - action: exec
    exec: ls -la
```
It is important to understand that this is one command to be executed and not a shell environment. This means shell functionality like `&&` or `|` are not working directly. Instead a bash has to be executed with the shell instructions. 

```yaml
x-direktiv:  
  cmds:
  - action: exec
    exec: bash -c 'ls -la | grep lib'
```

This logging of the command can be configured with the attributes `silent` and `print`. If `silent` is set to true the command does not log its output to Direktiv's log output. If `print` is set to true the command to be executed will be printed to Direktiv's logs. This is useful if the command uses secrets on the command line. Additionally environmment variables can be defined as well. For all attributes templating can be used as well.

```yaml
x-direktiv:  
  cmds:
  - action: exec
    exec: ls -la
    silent: true
    print: '{{ .ShouldPrint }}'`
    env: ["HELLO=world", "VALUE={{ .Hello }}"]
```

If a command generates a JSON output file it can be set as the response for this command using the `output` attribute. At the end of the command the action builder will read this file and uses it as the response for this command. 

```yaml
- action: exec
  exec: |-
    bash -c 'echo { \"hello\": \"{{ .Name }}\"  > /tmp/output }'
  output: /tmp/output
```

If a command produces an error it is recommended to use the `debug` setting in `x-direktiv` to see which templates and values are going into the templating function.

There are two additional attributes for basic error handling `continue` and `error`. If a command has set the `continue` value to true it will run the next command even if an error occurred. With the value `error` a custom error type can be thrown and handled in Direktiv. 

```yaml
- action: exec
  exec: this command fails
  error: my.small.error
  continue: true
```

### Execute (Foreach)

If the `action` attribute is set to `foreach` the defined command will be executed in a foreach loop. The attributes are identical to the single command. The only extra value which has to be provided is the value the command should loop over. It is called `loop` and has to point to an array of the incoming data. The following YAML shows a definition of an array of incoming data.

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
```

The payload of the incoming data can be the following:

```json
{
	"names": [
		"John",
		"Sarah",
		"Mike"
	]
}
```

The foreach can iterate through the names array and execute a command for each individual value in that list.

```yaml
x-direktiv:  
  debug: true
  cmds:
  - action: foreach
    loop: .Names
    exec: 'echo hello {{ .Item }}'
```

In foreach commands the object used in templating provides two values which can be used. One is `.Item` which is the actual object in the loop. The whole request object is in `.Body` if the templating values need access to values outside of the actual loop value. 

### HTTP Request

The `http` request action type is a simple method to do a request to backend systems and can be used to wrap exitsing APIs into easier-to-use Direktiv actions. For all attributes, except `errorNo200`, templating can be used.  

```yaml
- action: http
  url: http://www.direktiv.io/{{ .PathToExecute }}
  method: {{ .Method }}
  headers: 
    - Hello: {{ .World }}
    - Content-Type: application/text
  username: user
  password: mysecretPassword
  insecure: true 
  errorNo200: true
  continue: true
```

Additionally to the basic http atributes (url, method, headers), the http command provides additional attribues to set. It provides basic authentication via the `username` and `password` atributes and if the backend uses self-signed certificates the `insecure` flag can be set. The `errNo200` field defines if status codes larger that 299 should be treated as errors. 

For POST and PUT request there are two ways to send data to the backend service. It can be either defined in plain-text in a `data` attribute or in base64 for binary data. 

*Plain Text POST Example*
```yaml
x-direktiv:  
  debug: true
  cmds:
  - action: http
    url: http://www.direktiv.io
    method: post
    data: |
      {
        "greeting": "{{ .Name }}" }}"
      }
```

*Binary POST Example*
```yaml
x-direktiv:  
  debug: true
  cmds:
  - action: http
    url: http://www.direktiv.io
    method: post
    data64: amVuc2dlcmtl
```

### HTTP Request (Foreach)

The foreach functionality for http behaves exactly like the `foreach` for commands. It requires an additional `loop` attribute to define which value from the payload is used to iterate. 

```yaml
x-direktiv:  
  debug: true
  cmds:
  - action: foreachHttp
    loop: .Names
    url: http://www.direktiv.io/{{ .Item }}
    method: get
```

## Generating the action

After configuring the action the source code can be generated with the `gen` command. The command needs an additional argument which is the version number of the action to build. For the default init it is `v1.0.0.`. 

```
docker run -v `pwd`:/tmp/app direktiv/action-builder gen v1.0.0
```

After running this command the action is ready to be compiled, tested and deployed. The `run.sh` file in the folder should compile and run the application locally. After testing it can be pushed to a container registry and consumed by Direktiv. 

## Custom Go Code

The application builder is written in Go and the application can be written from scratch. To do this there is the `gen-custom` command which has to be called after the `init` command.

```
docker run -v `pwd`:/tmp/app direktiv/action-builder gen-custom v1.0.0
```

This generates a skeleton application where three functions can be implemented. Thes files containing those functions will never be overwritten with subsequent `gen-custom` calls.

### restapi/direktiv_post.go

This file contains the function `PostDirektivHandle` which is the main function getting executed from Direktiv. The parameter is a Go object which contains the information configured in the parameter section in the swagger.yaml file. 

This file contains a second function which is called before the Knative pod is getting destroyed. In this function cleanup or disconnecting routines can be added. 

### restapi/direktiv_delete.go

Direktiv sends a DELETE request if an action has been cancelled. It is up to the developer to implement this method. It provides the action ID.

## Generating documentation

TODO




