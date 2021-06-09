# Actions
Mokapi offers a set of built-in that you can use in expressions.

## find
Search for an element that matches the conditions defined by the specified predicate and returns the first occurrence within the sequence.

### Usage
```yaml
steps:
  - uses: echo
    with:
      msg: "${{ find([1,2,3,4], x => x == 3) }}"
```

## findAll
Retrieve all the elements that match the conditions defined by the specified predicate

### Usage
```yaml
steps:
  - uses: echo
    with:
      msg: "${{ findAll([1,2,3,4], x => x > 2) }}"
```

## any
Determines whether any element of a sequence matches the condition defined by the specified predicate

### Usage
```yaml
steps:
  - uses: echo
    with:
      msg: "${{ any([1,2,3,4], x => x == 2) }}"
```

## format
Formats according to a format specifier and returns the resulting string.

### Usage
```yaml
steps:
  - uses: echo
    with:
      msg: '${{ format("Hello {0}", "World") }}'
```

## randInt
Returns a non-negative pseudo-random int.

### Usage
```yaml
steps:
  - uses: echo
    with:
      msg: '${{ randInt() }}'
```

## randFloat
Returns a non-negative pseudo-random floating number in [0.0,1.0).

### Usage
```yaml
steps:
  - uses: echo
    with:
      msg: '${{ randInt() }}'
```