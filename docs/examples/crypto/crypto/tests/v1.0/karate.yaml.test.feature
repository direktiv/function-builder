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
	