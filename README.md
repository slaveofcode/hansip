
<img src="https://raw.github.com/slaveofcode/hansip/feature/configure-for-branding/assets/logo-256.png" align="right" />

# Hansip

Simple file sharing with End-to-End encryption for small/medium organization.

## Installation

### Requirements

1. Golang 1.18 or newer
2. Postgres 12 or newer

### Database

Before running any application or after preparing a new database, please execute the command below to add UUID extension. Otherwise the migration for table could faca an error because unsupported UUID function is executed.

#### Create UUID extension

> CREATE EXTENSION IF NOT EXISTS "uuid-ossp";