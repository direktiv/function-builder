# My Container

This is a short description.

---

## Description

This is a long description. This is a long description. This is a long description. This is a long description. This is a long description. This is a long description. This is a long description. This is a long description. This is a long description. This is a long description. This is a long description. This is a long description. This is a long description. This is a long description. This is a long description. This is a long description. This is a long description. This is a long description. This is a long description. This is a long description.

### Usage

```yaml
functions:
- id: aws
  image: direktiv/aws-cli
  type: knative-workflow
states:
- id: start
  type: action
  action:
    function: aws
    input: 
      access-key: MYACCESSKEY
      secret-key: sUpERSecrETKEY
      region: eu-central-1
      commands: 
        - ec2 create-security-group --group-name jq(.name) --description jq(.name)
          --tag-specifications ResourceType=security-group,Tags=[{Key=direktiv,Value=build},{Key=name,Value=jq(.name)}]
        - ec2 authorize-security-group-ingress --group-name jq(.name) --cidr 0.0.0.0/0 --protocol tcp --port 443
```

### Attributes

|Attribute|Description|
|---|---|
|access-key|Defines access key|
|secret-key|Defines secret key|
|region|select region|
|commands|List of commands to run|

### Errors

|Code|Reason|
|---|---|
|io.direktiv.compile1|Happens sometimes|
|io.direktiv.run|Happens sometimes as well|

### Example Request

```json
{
	"name": "jens"
}
```

### Example Response

```json
{
	"greeting": "Hello jens"
}
```

---

# NOT PART OF THE CONTENT

For the quick copy thing on the side we could use the function definition, e.g.

```yaml
- id: aws
  image: direktiv/aws-cli
  type: knative-workflow
```

Additional information we have and it might be good to show it seperatly:

- Tag / Version
- Update date (Last Updated)
- Category

