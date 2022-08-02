
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
