#!/bin/bash
set -e

exec tini -- oneblock webhook "${@}"
