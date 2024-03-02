
# Accounts service #

[![Go Report Card](https://goreportcard.com/badge/github.com/Falokut/accounts_service)](https://goreportcard.com/report/github.com/Falokut/accounts_service)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/Falokut/accounts_service)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Falokut/accounts_service)
[![Go](https://github.com/Falokut/accounts_service/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/Falokut/accounts_service/actions/workflows/go.yml) ![](https://changkun.de/urlstat?mode=github&repo=Falokut/accounts_service)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)

# Content

+ [About service](#about-service)
    + [Features](#features)
        + [Accounts and authentication](#accounts-and-authentication)
        + [Registration](#registration)
    + [Events](#events)
+ [Configuration](#configuration)
    + [Params info](#configuration-params-info)
        + [Database config](#database-config)
        + [Jaeger config](#jaeger-config)
        + [Prometheus config](#prometheus-config)
        + [time.Duration](#timeduration-yaml-supported-values)
+ [Metrics](#metrics)
+ [Docs](#docs)
+ [Author](#author)
+ [License](#license)
---------

# About service

The Account Service is a robust and secure service that provides essential functionalities for user accounts management. It offers a seamless user experience with features such as registration, password reset, account confirmation, login, and authentication.

# Features

1. Registration: Users can create new accounts by providing their basic information, including email and password. The registration process ensures that only valid and unique email addresses are accepted.

2. Password Reset: In case users forget their passwords, the service allows them to initiate a password reset procedure. A secure link is sent to the user's registered email address, enabling them to set a new password and regain access to their account.

3. Account Confirmation: To enhance security and prevent abuse, newly registered users must confirm their email addresses. A confirmation link is sent to the provided email, and upon verification, the account is activated within the system.

4. Login: Once registered and confirmed, users can securely log in to their accounts using their email and password. The service utilizes robust authentication protocols to protect account information and ensure secure access.

5. Authentication: To enhance security and prevent unauthorized access, the service employs authentication methods such as session-based identification and client identification based on their machine ID. If the machine ID provided in the request does not match the one stored in the session cache, access will be denied. These security measures ensure the safeguarding of user accounts and help in protecting against unauthorized access.

The Account Service provides a reliable, efficient, and user-friendly solution for managing user accounts in web applications. With its comprehensive set of features, it ensures the security and integrity of user data, delivering a seamless login and account management experience.

## Accounts and authentication
The accounts service features a login system where users can securely log in via sessions. This system ensures that only approved users can perform actions with their accounts.

To create an account, users can register by providing their email and password. Once registered and confirmed emails, users can log in to their accounts using their credentials. The system will generate a session token for the user, which they will use for authentication in future requests.

Users remaster logged in until they manually log out or their session expires. This eliminates the need for users to repeatedly authenticate themselves for each request, providing a seamless experience.

Users can safely access the services using their account information. Additionally, it's worth noting that passwords are encrypted and not stored in plain text. Instead, they are encrypted using encryption algorithm bcrypt. This provides an added layer of security, as even in the event of a data breach, it would be extremely difficult for malicious actors to recover and exploit these passwords.

When registering a new account, the entered passwords are securely encrypted before being stored in the database. This way, user passwords are protected from unauthorized access.

## Registration
During the registration process, an email confirmation link is sent to the user's provided email address (need another request). The user must click on this link to verify their account and activate it. Once the email is confirmed, the account information is securely transferred from the Redis cache to the master database.

Implementing this email verification step helps ensure that only legitimate users with valid email addresses can create accounts on the cinema ticket. It helps prevent potential abuse or unauthorized access by requiring users to verify their identities before gaining full access to the system.

---

# Events
The service generate 2 types of events: requests for the delivery of [tokens](./internal/events/tokensDeliveryMQ.go) to the user and events that occur with the [accounts](./internal/events/accountsEvents.go)(its creation, deletion, change of email). [events package](./internal/events/events.go)

---

# Configuration
1. Create .env in root dir  
Example env for redis:
```env
REDIS_PASSWORD=redispass
REDIS_AOF_ENABLED=no
```
2. [Configure accounts_db](accounts_db/README.md#Configuration)
3. Create a configuration file or change the config.yml file in docker\containers-configs.
If you are creating a new configuration file, specify the path to it in docker-compose volume section (your-path/config.yml:configs/)
4. Configure kafka broker [example compose file](kafka-cluster.yml)

## Configuration params info
if supported values is empty, then any type values are supported

| yml name | yml section | env name | param type| description | supported values |
|-|-|-|-|-|-|
| log_level   || LOG_LEVEL  |   string   |      logging level        | panic, fatal, error, warning, warn, info, debug, trace|
| profiles_service_addr   |      | PROFILES_SERVICE_ADDR  |    string       | ip address(or host) with port of profiles service            | all valid addresses formatted like host:port or ip-address:port |
| healthcheck_port   |      | HEALTHCHECK_PORT  |   string   |     port for healthcheck       | any valid port that is not occupied by other services. The string should not contain delimiters, only the port number|
| host   |  listen    | HOST  |   string   |  ip address or host to listen   |  |
| port   |  listen    | PORT  |   string   |  port to listen   | The string should not contain delimiters, only the port number |
| server_mode   |  listen    | SERVER_MODE  |   string   | Server listen mode, Rest API, gRPC or both | GRPC, REST, BOTH|
| allowed_headers   |  listen    |  |   []string, array of strings   | list of all allowed custom headers. Need for REST API gateway, list of metadata headers, hat are passed through the gateway into the service | any strings list|
| allowed_outgoing_header   |  listen    |  |   map[string]string  | map of headers, thath passess throught gateway from service (outgoing headers), which key is pretty header name, value is header name inside service | any map with string key and value string |
| service_name   |  prometheus    | PROMETHEUS_SERVICE_NAME | string |  service name, thats will show in prometheus  ||
| server_config   |  prometheus    |   | nested yml configuration  [metrics server config](#prometheus-config) | |
| nonactivated_account_ttl   |      |   | time.Duration with positive duration | the time that registered(non activated) account will be stored in the cache |[supported values](#time.Duration-yaml-supported-values)  |
| sessions_ttl   |      |   | time.Duration with positive duration | the time that session will be stored in the cache |[supported values](#time.Duration-yaml-supported-values)  |
|db_config|||nested yml configuration  [database config](#database-config) || configuration for database connection | |
|jaeger|||nested yml configuration  [jaeger config](#jaeger-config)|configuration for jaeger connection | |
| network  |  registration_repository |  REGISTRATION_REPOSITORY_NETWORK |  string |   | tcp or udp  |
| addr  |  registration_repository | REGISTRATION_REPOSITORY_ADDRESS  |string|ip address(or host) with port of redis| all valid addresses formatted like host:port or ip-address:port|
| password  |  registration_repository |  REGISTRATION_REPOSITORY_PASSWORD |  string | password for connection to the redis  |   |
|  db | registration_repository  | REGISTRATION_REPOSITORY_DATABASE  |  int | the number of the database in the redis  |   |
| network  |  sessions_repository |  SESSIONS_REPOSITORY_NETWORK |  string |   | tcp or udp  |
| addr  |  sessions_repository | SESSIONS_REPOSITORY_ADDRESS  |string|ip address(or host) with port of redis| all valid addresses formatted like host:port or ip-address:port|
| password  |  sessions_repository |  SESSIONS_REPOSITORY_PASSWORD |  string | password for connection to the redis  |   |
|  db | sessions_repository  | SESSIONS_REPOSITORY_DATABASE  |  int | the number of the database in the redis  |   |
|num_retries_for_terminate_sessions|||int|number of retries for session termination, when deleting account||
|retry_sleep_time_for_terminate_sessions||| time.Duration with positive duration | the time delay between session deletion retries|[supported values](#time.Duration-yaml-supported-values)|
|bcrypt_cost|crypto|BCRYPT_COST| int |the bcrypt hashing complexity|4-31|
| ttl  |  change_password_token |  | time.Duration with positive duration| the amount of time this token will be valid for|[supported values](#time.Duration-yaml-supported-values)|
| secret  |  change_password_token |  CHANGE_PASSWORD_TOKEN_SECRET |  string | the secret to generating a jwt token  ||
| ttl  |  verify_account_token |  | time.Duration with positive duration| the amount of time this token will be valid for|[supported values](#time.Duration-yaml-supported-values)|
| secret  |  verify_account_token |  VERIFY_ACCOUNT_TOKEN_SECRET |  string | the secret to generating a jwt token  ||
| ttl  |  change_email_token |  | time.Duration with positive duration| the amount of time this token will be valid for|[supported values](#time.Duration-yaml-supported-values)|
| secret  |  change_email_token |  CHANGE_PASSWORD_TOKEN_SECRET |  string | the secret to generating a jwt token  ||
| brokers  |  account_events |  |  []string, array of strings| list of the addresses of kafka brokers| any list of addresses like host:port or ip-address:port|
| brokers  |  tokens_delivery |  |  []string, array of strings| list of the addresses of kafka brokers| any list of addresses like host:port or ip-address:port|

### Database config
|yml name| env name|param type| description | supported values |
|-|-|-|-|-|
|host|DB_HOST|string|host or ip address of database| |
|port|DB_PORT|string|port of database| any valid port that is not occupied by other services. The string should not contain delimiters, only the port number|
|username|DB_USERNAME|string|username(role) in database||
|password|DB_PASSWORD|string|password for role in database||
|db_name|DB_NAME|string|database name (database instance)||
|ssl_mode|DB_SSL_MODE|string|enable or disable ssl mode for database connection|disabled or enabled|

### Kafka config
|yml name| env name|param type| description | supported values |
|-|-|-|-|-|
|brokers | |[]string, array of strings| list of the addresses of kafka brokers| any list of addresses like host:port or ip-address:port|
|topic||string| topic name| any topic name|

### Jaeger config

|yml name| env name|param type| description | supported values |
|-|-|-|-|-|
|address|JAEGER_ADDRESS|string|hip address(or host) with port of jaeger service| all valid addresses formatted like host:port or ip-address:port |
|service_name|JAEGER_SERVICE_NAME|string|service name, thats will show in jaeger in traces||
|log_spans|JAEGER_LOG_SPANS|bool|whether to enable log scans in jaeger for this service or not||

### Prometheus config
|yml name| env name|param type| description | supported values |
|-|-|-|-|-|
|host|METRIC_HOST|string|ip address or host to listen for prometheus service||
| port|METRIC_PORT|string|port to listen for  of prometheus service| any valid port that is not occupied by other services. The string should not contain delimiters, only the port number|

### time.Duration yaml supported values
A Duration value can be expressed in various formats, such as in seconds, minutes, hours, or even in nanoseconds. Here are some examples of valid Duration values:
- 5s represents a duration of 5 seconds.
- 1m30s represents a duration of 1 minute and 30 seconds.
- 2h represents a duration of 2 hours.
- 500ms represents a duration of 500 milliseconds.
- 100µs represents a duration of 100 microseconds.
- 10ns represents a duration of 10 nanoseconds.


# Metrics
The service uses Prometheus and Jaeger and supports distributed tracing

# Docs
[Swagger docs](swagger/docs/accounts_service_v1.swagger.json)

# Author

- [@Falokut](https://github.com/Falokut) - Primary author of the project

# License

This project is licensed under the terms of the [MIT License](https://opensource.org/licenses/MIT).

---