---
title: Declarative Data Generation
description: Providing a detailed specification for data types can greatly enhance the usefulness and accuracy of randomly generated data by ensuring that the generated data aligns with real-word scenarios and is consistent and error-free.
---
# Declarative Data Generation

Providing a detailed specification for data types can greatly enhance the usefulness and accuracy of randomly generated data by ensuring that the generated data aligns with real-word scenarios and is consistent and error-free.
Mokapi provides you an API endpoint to generate test data on the fly.

```Bash
curl -X POST http://localhost:8080/api/schema/example -H 'Content-Type: application/json' -d '{"type":"string","format":"date-time"}'
```

## Using format
OpenAPI and Mokapi provides some built-in string formats

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
      format: password # F6d?sZESB(3l
    email:
      type: string
      format: email # demetrisdach@yost.org
    guid:
      type: string
      format: uuid # dd5742d1-82ad-4d42-8960-cb21bd02f3e7
    url:
      type: string
      format: uri # http://www.leadvortals.name/global/aggregate/vertical/paradigms
    ipv4:
      type: string
      format: ipv4 # 187.211.129.91
    ipv6:
      type: string
      format: ipv6 # d36:6593:10e8:9818:19ce:a1f7:5f48:731a
    custom:
      type: string
      format: {firstname} {lastname} # Danny Robel
```

## Using pattern
With *pattern* you can describe the data with a regular expression.

```yaml
schema:
  type: object
  properties:
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
    user: 
      type: object
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
    user:
      type: object
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

Supported formats:

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
- {name}
- {firstname}
- {lastname}
- {gender}
- {email}

**Auth** <span id="auth"></span>
- {username}
- {password}

**Address** <span id="address"></span>
- {street}
- {city}
- {state}
- {zip}
- {country}
- {latitude}
- {longitude}

**Beer** <span id="beer"></span>
- {beername}
- {beeralcohol}
- {beerstyle}
- {beermalt}
- {beerhop}
- {beerblg}
- {beeribu}

**Car** <span id="car"></span>
- {carmaker}
- {carmodel}
- {cartype}
- {carfueltype}

**Words** <span id="words"></span>
- {noun}
- {verb}
- {adverb}
- {sentence}
- {paragraph}
- {loremipsumword}
- {loremipsumsentence}
- {loremipsumparagraph}
- {question}
- {phrase}

**Foods** <span id="foods"></span>
- {fruit}
- {vegetable}
- {breakfast}
- {lunch}
- {dinner}
- {snack}
- {dessert}

**Misc** <span id="misc"></span>
- {bool}
- {uuid}
- {flipacoin}

**Colors** <span id="colors"></span>
- {color}
- {hexcolor}
- {rgbcolor}
- {safecolor}

**Internet** <span id="internet"></span>
- {url}
- {domainname}
- {domainsuffix}
- {ipv4address}
- {ipv6address}
- {macaddress}
- {httpstatuscode}
- {loglevel}
- {httpmethod}
- {useragent}

**Date/Time** <span id="datetime"></span>
- {date},{date:UnixDate},{date:yy-dd-mm}  (Format: ANSIC, UnixDate, RubyDate, RFC822, RFC822Z, RFC850, RFC1123, RFC1123Z, RFC3339Nano, default: RFC3339)
- {nanosecond}
- {second}
- {minute}
- {hour}
- {day}
- {weekday}
- {year}

**App** <span id="app"></span>
- {appname}
- {appversion}
- {appauthor}

**Animals** <span id="animals"></span>
- {petname}
- {animal}
- {animaltype}
- {farmanimal}
- {cat}
- {dog}

**Emojis** <span id="emojis"></span>
- {emoji}
- {emojidescription}
- {emojicategory}
- {emojialias}
- {emojitag}

**Language** <span id="language"></span>
- {language}
- {programminglanguage}

**Numbers** <span id="numbers"></span>
- {number:min,max}, {number:1,10}
- {int8}
- {int16}
- {int32}
- {int64}
- {uint8}
- {uint16}
- {uint32}
- {uint64}
- {float32}
- {float32range:min,max}
- {float64}

**String** <span id="string"></span>
- {digit}
- {letter}
- {lexify}
- {nummerify}

**File** <span id="file"></span>
- {fileextension}
- {filemimetype}

**Payment** <span id="ayment"></span>
- {price:min,max}
- {currencyshort}
- {currencylong}
- {creditcardnumber}