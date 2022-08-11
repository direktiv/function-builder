# AWS Service

This is a more complex example with the AWS command line interface. It shows the use of environment variables, args and reponse values. This example will list all instance names in a specified region in AWS. 

# Prepare

```console
mkdir aws && cd aws
function-builder prepare -f aws
```

# Swagger.yaml

The first step is to figure out what variables are required for this service. The absolut minimum is an AWS access key (`access-key`) and secret key (`secret-key`). The third value is the region because the service should be able to query different regions. The access and secret key are required and we will use a defult region if it is not set. 

```yaml
- name: body
  in: body
  schema:
    type: object
    required:
      - access-key
      - secret-key
    properties:
      access-key:
        type: string
      secret-key:
        type: string
      region:
        type: string
```

# Dockerfile 

We know that we want to run the AWS cli so it is necessary to change the `Dockerfile` to use that image. For this we just need to change the `FROM` statement in the generated Dockerfile:


```dockerfile
FROM golang:1.18.2-alpine as build

WORKDIR /src

COPY build/app/go.mod go.mod
COPY build/app/go.sum go.sum

RUN go mod download

COPY build/app/cmd cmd/
COPY build/app/models models/
COPY build/app/restapi restapi/

ENV CGO_LDFLAGS "-static -w -s"

RUN go build -tags osusergo,netgo -o /application cmd/aws-server/main.go; 

FROM amazon/aws-cli:2.7.21

COPY --from=build /application /bin/application

EXPOSE 8080

# remove image endpoint
ENTRYPOINT []

CMD ["/bin/application", "--port=8080", "--host=0.0.0.0", "--write-timeout=0"]
```

# Command

The next step is to configure the command. The basic command is the [describe-instances](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-instances.html) command. 

```
aws ec2 describe-instances
```

This request requires authentication so we need to provide these details as well as the region. This can be done with [environment variables](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html). 

- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY
- AWS_DEFAULT_REGION

All this information translates to the following command in `swagger.yaml`:

```yaml
x-direktiv:
cmds:
- action: exec
    exec: aws ec2 describe-instances --query "Reservations[].Instances[].InstanceId"
    env:  
    - AWS_ACCESS_KEY_ID={{ .AccessKey }}
    - AWS_SECRET_ACCESS_KEY={{ .SecretKey }}
    - AWS_DEFAULT_OUTPUT=json
    - AWS_DEFAULT_REGION={{ default "us-east-1" .Region }}
```

The interesting part is the `AWS_DEFAULT_REGION` environment variable. The template command `default` uses the value `us-east-1` if no region is set in the request. 

For testing we need to configure a few more things. First of all we define the required secrets. Defining secrets in `swagger.yaml` is iportant for two things: It generates this information for the auto-generated documentation and it modifies the test scripts in the folder `test` to use and request those secrets. The screct section could look like the following:

```yaml
x-direktiv-secrets:
  - name: awsKey
    description: This is the AWS key
  - name: awsSecret
    description: This is the AWS secret
```

We change the response to allow every response also.

```yaml
responses:
  200:
    description: list of instances
    schema:
      type: object
```

Now it is time to generate the source code and run it.

```console
function-builder gen
./run.sh
```

The function can be tested with the `run-tests.sh` script. Because there are secrets defined they need to be provided for testing. During the `gen` phase the test script has been modified to require the two configured secrets. They can be passed to the test on the command line. 

```console
DIREKTIV_SECRET_awsKey=MYACC377K3Y DIREKTIV_SECRET_awsSecret=MyAwSs3Cr3t ./run-tests.sh
```

These variables are used in the Karate test script in `tests/karate.yaml.test.feature`.

```
Feature: Basic

# The secrects can be used in the payload with the following syntax #(mysecretname)
Background:
* def awsKey = karate.properties['awsKey']
* def awsSecret = karate.properties['awsSecret']


Scenario: aws

	Given url karate.properties['testURL']

	And path '/'
	And header Direktiv-ActionID = 'development'
	And header Direktiv-TempDir = '/tmp'
	And request
	"""
	{
		"access-key": #(awsKey),
		"secret-key": #(awsSecret),
	}
	"""
	When method POST
	Then status 200
	And match $ ==
	"""
	[
	{
		"result": #notnull,
		"success": true
	}
	]
	"""
	
```







