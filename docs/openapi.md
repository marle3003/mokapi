# OpenAPI
Mokapi supports OpenAPI 3.0. With [Enum](https://swagger.io/docs/specification/data-models/enums/) and [Example](https://swagger.io/docs/specification/adding-examples/) you can define the responses of your mocked Services without any coding.
Mokapi takes a random element of your *enumeration* or the defined *example* element to response to a request.
Additionally, it is possible to define a random generator for your data.

## Enum

```yaml
schema:
  type: object
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

```yaml
schema:
  type: object
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
```

Supported generator:

**Person**
- name
- firstname
- lastname
- gender
- email

**Auth**
- username
- password

**Beer**
- beername
- beeralcohol
- beerstyle
- beermalt
- beerhop
- beerblg
- beeribu

**Car**
- carmaker
- carmodel
- cartype
- carfueltype

**Words**
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

**Foods**
- fruit
- vegetable
- breakfast
- lunch
- dinner
- snack
- dessert

**Misc**
- bool
- flipacoin

**Colors**
- color
- hexcolor
- rgbcolor
- safecolor

**Internet**
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

**Date/Time**
- date
- nanosecond
- second
- minute
- hour
- day
- weekday
- year

**App**
- appname
- appversion
- appauthor

**Animal**
- petname
- animal
- animaltype
- farmanimal
- cat
- dog

**Emoji**
- emoji
- emojidescription
- emojicategory
- emojialias
- emojitag

**Language**
- language
- programminglanguage

**Number**
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

**String**
- digit
- letter
- lexify
- nummerify

**File**
- fileextension
- filemimetype