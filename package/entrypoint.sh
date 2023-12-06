#!/bin/bash
set -e

exec tini -- oneblock api-server "${@}"
