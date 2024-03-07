---
title: Contract Testing - An Introduction
description: Take a closer look at why you can achieve better time to market with contract testing.
---
# Contract Testing

Contract testing gets increasingly more crucial with a micro service architecture. Especially in an agile development team to be able to test interactions between service as the system changes quickly in the production. Contract testing ensures that the *contract* or agreement between a service and its consumers remains intact and unbroken.

This type of testing allows us to identify discrepancies between expected and actual data formats and response codes and thereby preventing integration issues. By automating these checks, agile teams can detect problems early in the development cycle, reducing the risk of production outages and facilitating smoother deployments.

Integrating a mock tool like Mokapi into the contract testing process can significantly reduce an organizations's time to market for several reasons:

1. **Faster Feedback Loops**: Mock tools allow developers to simulate the responses of various services without having to set up and run those services in a live environment. This means developers can quickly test and iterate on their code, receiving immediate feedback on any issues that need to be addressed. Faster feedback loops translate to quicker resolution of problems and a more agile development process.
2. **Parallel Development**: With mock tools, frontend and backend teams or teams for different services can work simultaneously rather than sequentially. While one team is developing the actual services, the other team can use mocked responses based on the contract to continue their work. This parallel development process eliminates bottlenecks where one team is waiting for the other to complete their part, speeding up the overall development timeline and therefore time to market.
3. **Reduced Infrastructure Costs**: By using mock tools, the need for a fully deployed and integrated environment for early-stage testing is reduced. Teams can test against mocked services without the overhead of setting up and maintaining a complete stack of services. This is further enhanced if the contract of an API can be integrated automatically from a directory. This not only saves on infrastructure costs but also on the time required to manage these environments.
4. **Improved Reliability**: Mock tools help enforce contract testing by simulating services based on the agreed-upon contract. This ensures that any deviation from the contract is caught early in the development cycle. By identifying discrepancies early, problems can be resolved before they reach production.
5. **Enhanced Focus on User Experience**: Since mock tools can simulate various responses, including edge cases, errors, and failure modes, developers can design and test how the application behaves under different conditions. This allows for a more thorough exploration of potential user scenarios, ensuring that the end product is not only functional but also delivers a superior user experience.

In summary, incorporating mock tools into the contract testing strategy enhances the development process by enabling faster feedback, parallel development, and reduced infrastructure requirements. This holistic improvement in efficiency and effectiveness directly contributes to an organization's ability to bring products to market faster, giving it a competitive advantage in today's fast-moving digital landscape.