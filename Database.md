# Database structure

## SignKeys
  - id
  - key
  - expiry

## EncKeys
  - id
  - key
  - expiry

## Users
  - id
  - displayName
  - created
  - isDeleted

## UserConnections
  - userId
  - signInType
  - accountId
  - authDetails

## Fellowships
  - id
  - name
  - creator

## FellowshipMembers
  - fellowshipId
  - userId
  - access

## FellowshipCircles
  - id
  - fellowshipId
  - name
  - type
  - creator

## CircleMembers
  - circleId
  - userId
  - access

## Posts
  - id
  - authorId
  - fellowshipId
  - circleId
  - posted
  - heading
  - article
