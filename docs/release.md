# Mokapi Release 0.32.2-alpha

⚠️ Automatically generated alpha build.

- Alpha version: 0.32.2-alpha
- Commit: e874ee62
- Built at: 2026-02-01 18:17:56 UTC

## Changes since v0.32.1
    
- improve git log messages
- add remove author filter
- fix remove author filter
- fix parameter name
- fix syntax error
- remove git log filter
- fix git log filter remove debug output
- add debug output
- fix fetch main branch
- fix fetch all tags
- fix parameter change version format for alpha
- fix parameter
- add build alpha release notes for dashboard
- refactor test books.spec.ts change HTTP service view in dashboard add metrics to HTTP operations table
- fix test
- add Kafka server bindings to dashboard
- fix url
- improve server view in dashboard add a default Kafka server if no server is defined in spec update doc improve Kafka client if no message or operation is defined for channel add Kafka examples
- fix tests
- fix Kafka protocol after changed to simplified leader management change broker selection only by port fix getting events by traits
- improve GIT clone performance
- fix precedence of CLI flags
- improve error message
- fix CLI setting value in correct order fix CLI setting slices
- fix Kafka API if group does not yet have a generation
- fix logging Kafka FindCoordinator add test
- Bump vue-router from 4.6.4 to 5.0.0 in /webui
- Bump nodemailer and @types/nodemailer in /webui
- Bump vue-tsc from 3.2.3 to 3.2.4 in /webui
- Bump @types/node from 25.0.10 to 25.1.0 in /webui
- add error handling to Kafka request logging
- add console logging fix dashboard invoke info request
- add logging Kafka InitProducerId
- add logging FindCoordinator improve logging ListOffsets
- improve test stability
- add renaming kafka ListOffsets
- add ListOffsets logging
- fix data loading
- fix data loading in dashboard add sync group logging
- add join group response
- fix displaying Kafka messages in client view
- add logging Kafka request JoinGroup and displaying in dashboard (WIP)
- improve tab click event improve dashboard to prevent the accidental loading of demo data
- fix concurrent map read/write
- add support of channel tags
- change view of kafka cluster for better UX
- adjust test
- add producer's clientId to kafka message log update Kafka message view in dashboard
- fix add old style of index format in dynamic flags fix add missing flags for GitHub auth
- fix CLI reading config file
- fix CLI reading config file fix responsiveness in dashboard
- fix responsiveness
- fix remove dynamic doc entry
- change Leader handling simplified: - Each partition and group now always reports the current server as leader/coordinator.
- improve Kafka dashboard for groups and clients
- change Kafka group displaying in dashboard
- fix typo in names
- remove npm token
