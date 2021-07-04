# Specifying response payload
Mokapi allow you to customize the response payload to meet the unique needs of your
application. In this guide, we'll discuss some essential customization techniques such
as using *format*, *pattern*, *enum*, *example*, *fake* and generate data by your own script. 

## Using format
Mokapi offers a set of built-in string format.

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

## Using pattern
With *pattern* you can describe the data with a regular expression.

```yaml
schema:
  type: object
    ssn:
      type: string
      pattern: '^\d{3}-\d{2}-\d{4}$' # 123-45-6789
```

## Using enum
You can use the *enum* keyword to specify possible values.

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

## Using example
With example, you can set the values for an object, array or property. Mokapi takes
exactly this value.

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
        email: # not present in payload
          type: string
          format: email
      example:
        - id: 11
          name: Jessica
```

## Using generator
The *x-faker* keyword lets you define a random data generator.

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

**Address** <span id="address"></span>
- street
- city
- state
- zip
- country
- latitude
- longitude

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
- uuid
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

**Payment** <span id="ayment"></span>
- price:min,max
- currencyshort
- currencylong
- creditcardnumber