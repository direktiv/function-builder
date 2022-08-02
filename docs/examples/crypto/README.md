# Cryptocurrency

In this example we want to build a simple crypto function based on an open web service provided by https://api.blockchain.com. We will use the http action to implement this functionality. 
This function acepts a symbol, e.g. BTC-USD or NEAR-EUR and returns trading information about the currency.

# Prepare

```console
mkdir crypto && cd crypto
function-builder prepare -f crypto
```

# Swagger.yaml

The first thing to change when creating a new function is the input. This is defined in the parameter section under `body`. In this case we want one parameter `symbol` and it is a mandatory parameter because the function would not work without it. 

```yaml
- name: body
  in: body
  schema:
    type: object
    required: ["symbol"]
    properties:
      symbol:
        type: string
```

After that change we remove `x-direktiv-secrets` because this function does not require any secret values. The next step is to define the command to execute. As we have already mentioned we are using the acion type `http` to access the blockchain API. The URL is the blochain API and we are using the variable `{{ .Symbol }}` at the end. That references the required parameter we specified in the input section and because the default http action returns headers too we are defining an output. Firt we want to get the output from the first command with `index . 0`. This containes the response of the http requests. With `index (index . 0) "result"` we are getting the `result` attribute from the http response and we want JSON with `toJson`. 


```yaml
x-direktiv:
  cmds:
  - action: http
    url: https://api.blockchain.com/v3/exchange/tickers/{{ .Symbol }}
  output: |
    { "crypto": {{ index (index . 0) "result" | toJson }} }
```

The next step is define the response. In this case we make it rather simple and allow every response. In real life functions you would define the response so you know the response have a certain layout.

```yaml
responses:
  200:
    description: Value of the crypto
    examples:
      crypto:
        crypto: 
          last_trade_price: 23039.44
          price_24h: 23317.82
          symbol: BTC-USD
          volume_24h: 391.2400001
    schema:
      type: object
```

# Testing

After this change we can test the function. For tests there is a `run.sh` file in the base directory. This script build sthe function and starts the docker container at port 9191. After starting the function we can run the tests with a helper script `run-tests.sh`. This script starts a container with the Karate API testing framework installed and it executes the tests configured in the file `tests\karate.yaml.test.feature`. The default looks like this:

```sh
Feature: Basic

# The secrects can be used in the payload with the following syntax #(mysecretname)
Background:


Scenario: get request

	Given url karate.properties['testURL']

	And path '/'
	And header Direktiv-ActionID = 'development'
	And header Direktiv-TempDir = '/tmp'
	And request
	"""
	{
		"commands": [
		{
			"command": "ls -la",
			"silent": true,
			"print": false,
		}
		]
	}
	"""
	When method POST
	Then status 200
	And match $ ==
	"""
	{
	"crypto": [
	{
		"result": "#notnull",
		"success": true
	}
	]
	}
	"""
```

Running the test fails with `symbol in body is required` and that is what we have configured earlier. Lets change the test file to send the symbol and test for the right response. This test wil run successfully and our crypto funciton is ready to be released. To generate the docs perfectly it still requires to add exmanples and defaults in the `swagger.yaml` file but the functionality is ready.


```sh
Feature: Basic

Scenario: crypto

	Given url karate.properties['testURL']

	And path '/'
	And header Direktiv-ActionID = 'development'
	And header Direktiv-TempDir = '/tmp'
	And request
	"""
	{
    "symbol": "BTC-USD"
	}
	"""
	When method POST
	Then status 200
	And match $ ==
	"""
	{
		"crypto": {
			"last_trade_price": #number,
    		"price_24h": #number,
    		"symbol": "BTC-USD",
    		"volume_24h": #number
		}
	}
	"""
```



