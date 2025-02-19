@fleet_server
Feature: Fleet Server
  Scenarios for Fleet Server, where an Elasticseach and a Kibana instances are already provisioned,
  so that the Agent is able to communicate with them

@start-fleet-server
Scenario Outline: Deploying an <os> Elastic Agent that starts Fleet Server
  When a "<os>" agent is deployed to Fleet with "tar" installer in fleet-server mode
  Then the agent is listed in Fleet as "online"

@centos
Examples: Centos
| os     |
| centos |

@debian
Examples: Debian
| os     |
| debian |
