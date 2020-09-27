#!/bin/bash

DB=${DB:-postgres}
TRACING=${TRACING:-false}
PROMETHEUS=${PROMETHEUS:-false}

DC="docker-compose -f start.yml up --build"

$DC

