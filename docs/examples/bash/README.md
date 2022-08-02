# Bash

In this example we want to build a function which searches and counts words in a webpage. 

# Prepare

```console
mkdir bash && cd bash
function-builder prepare -f bash
```

# Swagger.yaml

Because this service will be implemented in `bash`, we'll utilize three commands to count word occurrences on a webpage with curl. The first command is curl to get actually get it, and the other ones are grep and wc to do the actual searching and counting. But, before we go any further, let's take a look at the input. We need to accept two arguments. An `address` and a `search` phrase. Both are string values, and they are required to make this service operate. We need to edit the body input configuration to read as follows:


```yaml
- name: body
  in: body
  schema:
    type: object
    required:
      - address
      - search
    properties:
      address:
        type: string
      search:
        type: string
```

To achieve this functionality we use `curl` to get the webpage and `grep` the search string. At the end we are word counting (`wc`) the results. That translates into this command:

```yaml
exec: |-
    bash -c 'curl -sL {{ .Address }} | grep -o -i {{ .Search }} | wc -l'
```

As you can see we are using [go templating](https://pkg.go.dev/text/template) here. We need upper-case variables to reference the parameters from the input. This template means we are replacing `{{ .Address }}` with the incoming value as well as `search` in the `grep` instruction. For now we don't want to change the output to something custom so we can comment that out for now. The `x-direktiv` section should look like this now:

```yaml
x-direktiv:   
  cmds:
  - action: exec
    exec: |-
      bash -c 'curl -sL {{ .Address }} | grep -c {{ .Search }}'
```

The are two more little changes to make our service testing ready. The secret section `x-direktiv-secrets` has to be removed because this function doesn't required any. The second change is to modify the response to basically allow all responses. This is usually used to make sure the response has a certain layout. In this case we want to allow everything.

```yaml
responses:
  200:
    description: List of executed commands.
    schema:
      type: object
```
This function is ready for testing now. For this we need to modify the file `tests/karate.yaml.test.functions` to provide an address and a search term. The test should look like this:

```sh

Feature: Basic

Scenario: wordcount

	Given url karate.properties['testURL']

	And path '/'
	And header Direktiv-ActionID = 'development'
	And header Direktiv-TempDir = '/tmp'
	And request
	"""
	{
		"address": "https://www.google.com",
		"search": "search"
	}
	"""
	When method POST
	Then status 200
```

The code is now ready to be generated and run. The procedure for generating the code is straightforward, and it only takes one command to execute in the folder where the prepare command was run.

```
function-builder gen
```

After that we can start the function with the `run.sh` script and run the test with `run-tests.sh`.  This execution will fail and in the logs of the running service container we can see the error message:

```
bash: line 1: curl: command not found
```

This can be fixed by changing the generated Dockerfile to add the installation of `curl` and `grep`.


```Dockerfile
...

FROM ubuntu:22.04

RUN apt-get update && apt-get install ca-certificates curl grep -y

# DON'T CHANGE BELOW 
COPY --from=build /application /bin/application

EXPOSE 8080

CMD ["/bin/application", "--port=8080", "--host=0.0.0.0", "--write-timeout=0"]
```

If we run the service again with `run.sh` and executing a test again with `run-tests.sh` it will return with a successful response.

```json
[
	{
		"result": 6,
		"success": true
	}
]
```

Our service is working but we want to make the output a little bit more readable. For this we can change the `output` we removed before and change the name. The output is using a template as well and it is actually quite simple. Without an output specified the service returns an array of the results of the commands it has executed. In this case it was only one. To get the first item in the list we use `(index . 0)` and from this first item we want the `result` value: `{{ index (index . 0) "result" }}`. The final swagger.yaml file can be seen [here](bash/swagger.yaml).

```yaml
x-direktiv:   
    cmds:
    - action: exec
        exec: |-
        bash -c 'curl -sL {{ .Address }} | grep -c {{ .Search }}'
    output: |
      {
        "hits": "{{ index (index . 0) "result" }}"
      }
```
