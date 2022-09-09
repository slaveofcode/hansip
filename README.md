
<img src="https://raw.github.com/slaveofcode/hansip/develop/assets/logo-256.png" align="right" />

# Hansip

Simple file sharing with End-to-End encryption for small/medium organization.

### Server Requirements

1. Golang 1.18 or newer
2. Postgres 12 or newer

### Web Client

Hansip server will need [Hansip Web](https://github.com/slaveofcode/hansip-web) to interact with the users. It's a static site that can be deployed on static-site hosting server.

<h4 align="center">Home Page</h4>

![](https://raw.github.com/slaveofcode/hansip/develop/assets/screenshots/homepage.png)

<h4 align="center">Upload Preview</h4>

![](https://raw.github.com/slaveofcode/hansip/develop/assets/screenshots/upload-preview.png)

<h4 align="center">Extra Password Protection</h4>

![](https://raw.github.com/slaveofcode/hansip/develop/assets/screenshots/security-password.png)

## Installation
### Database

Before running any application or after preparing a new database, please execute the command below to add UUID extension. Otherwise the migration for table could face an error because unsupported UUID function is executed.

#### Create UUID extension

> CREATE EXTENSION IF NOT EXISTS "uuid-ossp";