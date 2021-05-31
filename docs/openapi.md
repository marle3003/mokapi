# OpenAPI
Mokapi generates random data for the given response schema. You have several configuration options to cover your needs. If you want your data to be dynamic, you should take a look at pipelines, e.g. data depends on request parameters.

## Format
```yaml
schema:
  type: object
  properties:
    date:
      type: string
      format: date # 2017-07-21
    time:
      type: string
      format: date-time # 2017-07-21T17:32:28Z
    password:
      type: string
      format: password
    email:
      type: string
      format: email
    guid:
      type: string
      format: uuid
    url:
      type: string
      format: uri
    ipv4:
      type: string
      format: ipv4
    ipv6:
      type: string
      format: ipv6
```

### Pattern
```yaml
schema:
  type: object
    ssn:
      type: string
      pattern: '^\d{3}-\d{2}-\d{4}$' # 123-45-6789
```

## Enum
Mokapi will take a random element from your [Enum](https://swagger.io/docs/specification/data-models/enums/) definition.

```yaml
schema:
  type: object
  properties:
    color:
      type: string
      enum: [red, green, blue]
    user: object
      properties:
        id:
          type: integer
        name:
          type: string
      enum:
        - id: 10
          name: Bob
        - id: 11
          name: Jessica
```

## Example
You can add [Example](https://swagger.io/docs/specification/adding-examples/) to properties, objects or arrays. Mokapi will take your defined example to generate the response. Additional defined elements are ignored.

```yaml
schema:
  type: object
  properties:
    color:
      type: string
      example: red
    user: object
      properties:
        id:
          type: integer
        name:
          type: string
      example:
        - id: 11
          name: Jessica
```

## Random generator
Mokapi extends the OpenAPI specification *Schema* object with a custom property *x-faker*. 

```yaml
schema:
  type: object
  properties:
    username:
      type: string
      x-faker: username
    count:
      type: integer
      x-faker: number:1,10
    custom:
      type: string
      x-faker: '{number:1,3} {beername}, {number:3,5} {fruit}'
    date:
      type: string
      x-faker: '{year}-{month}-{day}'
    dateUnix:
      type: string
      x-faker: date:UnixDate
```

Supported generator:

- [Person](#person)
- [Auth](#auth)
- [Beer](#beer)
- [Car](#car)
- [Words](#words)
- [Foods](#foods)
- [Misc](#misc)
- [Colors](#colors)
- [Internet](#internet)
- [Date/Time](#datetime)
- [App](#app)
- [Animals](#animals)
- [Emojis](#emojis)
- [Language](#language)
- [Numbers](#numbers)
- [String](#string)
- [File](#file)

**Person** <span id="person"></span>
- name
- firstname
- lastname
- gender
- email

**Auth** <span id="auth"></span>
- username
- password

**Beer** <span id="beer"></span>
- beername
- beeralcohol
- beerstyle
- beermalt
- beerhop
- beerblg
- beeribu

**Car** <span id="car"></span>
- carmaker
- carmodel
- cartype
- carfueltype

**Words** <span id="words"></span>
- noun
- verb
- adverb
- sentence
- paragraph
- loremipsumword
- loremipsumsentence
- loremipsumparagraph
- question
- phrase

**Foods** <span id="foods"></span>
- fruit
- vegetable
- breakfast
- lunch
- dinner
- snack
- dessert

**Misc** <span id="misc"></span>
- bool
- flipacoin

**Colors** <span id="colors"></span>
- color
- hexcolor
- rgbcolor
- safecolor

**Internet** <span id="internet"></span>
- url
- domainname
- domainsuffix
- ipv4address
- ipv6address
- macaddress
- httpstatuscode
- loglevel
- httpmethod
- useragent

**Date/Time** <span id="datetime"></span>
- date (Format: ANSIC, UnixDate, RubyDate, RFC822, RFC822Z, RFC850, RFC1123, RFC1123Z, RFC3339Nano, default: RFC3339)
- nanosecond
- second
- minute
- hour
- day
- weekday
- year

**App** <span id="app"></span>
- appname
- appversion
- appauthor

**Animals** <span id="animals"></span>
- petname
- animal
- animaltype
- farmanimal
- cat
- dog

**Emojis** <span id="emojis"></span>
- emoji
- emojidescription
- emojicategory
- emojialias
- emojitag

**Language** <span id="language"></span>
- language
- programminglanguage

**Numbers** <span id="numbers"></span>
- number:min,max
- int8
- int16
- int32
- int64
- uint8
- uint16
- uint32
- uint64
- float32
- float32range:min,max
- float64

**String** <span id="string"></span>
- digit
- letter
- lexify
- nummerify

**File** <span id="file"></span>
- fileextension
- filemimetype