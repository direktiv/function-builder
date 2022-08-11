
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
	