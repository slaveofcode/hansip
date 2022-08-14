# Securi

Simple & protected file sharing for small organization

## DB Preparation

Before running any application or after preparing a new database, please execute the command below to add UUID extension. Otherwise the migration for table could faca an error because unsupported UUID function is executed.

### Create UUID extension

> CREATE EXTENSION IF NOT EXISTS "uuid-ossp";