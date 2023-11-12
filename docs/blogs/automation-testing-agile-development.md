---
title: Automation Testing in Agile Development
description: Automated tests are essential for Agile or CI teams. Learn how Mokapi helps to build better software faster
---
# Automation Testing in Agile Development

Automated tests are essential for agile development. They ensure that our software system is still 
easy to change. In an agile development process, software systems are changed and new feature are 
added in many short iteration. Without an automated testing setup it becomes a challenge to keep up 
high quality of our product.

The longer a product is in development, the harder it becomes for manual testing to keep up with the 
pace. Manual testing does not scale, and it's not compatible with continuous delivery. In continuous delivery
software code is being developed in smaller increments, released more often and deployed more frequently. 
Imagine we have 50 production changes every day. For that, we need continuous testing to make sure that new 
code is bug-free before it hits the production environment. 

Automated testing such as unit testing and contract testing eliminates the fear of change.

> Write tests until fear is transformed into boredom
> <span>Kent Beck</span>

## Why Contract Testing?

<img src="/e2e-testing.png" width="700" alt="Contract Testing" title="Contract Testing" />

Our software systems are becoming more and more complex and need to be integrated with many 
external software systems that are maintained by other teams. We can't test our software system including 
all external systems. We would need to understand all these external systems (system A and B) and even get 
them in the right state for our tests. Also, these systems are changing, and we don't know how. Our tests 
would be dependent on external systems in relation to runtime, changes and bugs.

## Test only system that we are responsible for and fake all others

<img src="/systemtest.png" width="700" alt="End-to-End Testing" title="End-to-End Testing" />

For stable tests, we need to simulate all interactions with external systems. This way, we only test 
what we are responsible for, and thus have a much more controlled testing strategy. Mokapi can help us 
to simulate interfaces like REST API or Apache Kafka and allows us to bring the really important information
into our system under test. 

This can help us to improve the speed, accuracy, and efficiency of testing. The 
software development team can identify and fix issues faster and ultimately deliver higher quality software
to the end user.

<img src="/betterfaster.png" width="300" alt="Testing in Agile Development" title="Testing in Agile Development" style="text-align: center;display: block;" />