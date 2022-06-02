Feature: greeting end-point

Background:
* url demoBaseUrl
* string tmp = read('data/example.dat')

Scenario: say my name

    Given path '/'
    Given header Direktiv-ActionID = 'development'
    Given header Direktiv-Tempdir = '/tmp'
    And request { "name": "#(tmp)" }
    When method post
    Then status 200