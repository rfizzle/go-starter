#!/usr/bin/env bash

# Setup colors
RED=''
GREEN=''
YELLOW=''
BLUE=''
MAGENTA=''
CYAN=''
WHITE=''
RESET=''

# check if stdout is a terminal...
if test -t 1; then
    # see if it supports colors...
    if test -n "$(tput colors)" && test "$(tput colors)" -ge 8; then
      RED='\033[0;31m'
      GREEN='\033[0;32m'
      YELLOW='\033[0;33m'
      BLUE='\033[0;34m'
      MAGENTA='\033[0;35m'
      CYAN='\033[0;36m'
      WHITE='\033[0;37m'
      RESET='\033[0m'
    fi
fi