# AWS Service

# Init

```
docker run --user 1000:1000 -v `pwd`:/tmp/app direktiv/service-builder init instance-list
```

# Service

This is a more complex example with the AWS command line interface. It shows the use of environment variables, args and reponse values. This example will list all instance names in a specified region in AWS. 

# Input

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
FROM golang:1.18-buster as build

COPY go.mod src/go.mod
COPY go.sum src/go.sum
RUN cd src/ && go mod download

COPY cmd src/cmd/
COPY models src/models/
COPY restapi src/restapi/

RUN cd src && \
    export CGO_LDFLAGS="-static -w -s" && \
    go build -tags osusergo,netgo -o /application cmd/aws-cli-server/main.go; 

# change to aws image
FROM amazon/aws-cli:2.6.1


COPY --from=build /application /bin/application

EXPOSE 8080

# remove image endpoint
ENTRYPOINT []

CMD ["/bin/application", "--port=8080", "--host=0.0.0.0"]
```

# Command

The next step is to configure the command. The basic command is the [describe-instances](https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-instances.html) command. 

```
aws ec2 describe-instances
```

This request needs authentication so we need to provide these details as well as the  region. This can be done with [environment variables](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html). 

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

The interesting part is the `AWS_DEFAULT_REGION` environment variable. The template command `default` uses the value `us-east-1` if no region is set in the request. If we generate the service with `docker run --user 1000:1000  -v `pwd`:/tmp/app direktiv/service-builder gen v1.0.0` and run it with `run.sh` something like this back:

```json
[
	{
		"result": [
			"i-ababcabcabcabcabc",
			"i-ababcabcabcabcabc"
		],
		"success": true
	}
]
```

The response is the generic reponse for a command. We can use the `output` attribute to modify the response to be more service related:

```yaml
x-direktiv:
    debug: true  
    cmds:
    - action: exec
        exec: aws ec2 describe-instances --query "Reservations[].Instances[].InstanceId"
        env:  
        - AWS_ACCESS_KEY_ID={{ .AccessKey }}
        - AWS_SECRET_ACCESS_KEY={{ .SecretKey }}
        - AWS_DEFAULT_REGION={{ default "us-east-1" .Region }}
    output: |-
          { "instances": {{ index (index . 0) "result"  | toJson }} }
```

This output template gets the result of the first command which is the `describe-instances` command. The `toJson` instruction is to convert the object to JSON because the `output` attribute needs to be configured as JSON. 

This service is now ready to be used and it can be pushed to a container registry and used in Direktiv.

```yaml
functions:
- id: inst
  image: localhost:5000/instances
  type: reusable
states:
- id: inst 
  type: action
  action:
    secrets: ["aws-key", "aws-secret"]
    function: get
    input: 
      access-key: jq(.secrets."aws-key")
      secret-key: jq(.secrets."aws-secret")
```