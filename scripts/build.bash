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

# Command line variable for bash script
for arg in "$@"; do
  shift
  case "$arg" in
  "--linux") set -- "$@" "-l" ;;
  "--osx") set -- "$@" "-o" ;;
  "--windows") set -- "$@" "-w" ;;
  "--zip") set -- "$@" "-a" ;;
  "--verbose") set -- "$@" "-v" ;;
  "--clean") set -- "$@" "-c" ;;
  *) ;;
  esac
done

while getopts "lowavc" opt; do
  case "$opt" in
  l) LINUX=1 ;;
  o) OSX=1 ;;
  w) WINDOWS=1 ;;
  a) ZIP=1 ;;
  v) VERBOSE=1 ;;
  c) CLEAN=1 ;;
  *) echo -e "${RED}unknown option $opt${RESET}" ;;
  esac
done
shift $((OPTIND - 1))

# Get build and project directories
BUILD_PATH="$(realpath "$0")"
DIR_PATH="$(dirname "$BUILD_PATH")"
PROJECT_PATH=$(realpath "$DIR_PATH/../")
APPLICATION_NAME="starter"

# Change to project directory
cd "$PROJECT_PATH" || {
  echo -e "${RED}failed to get project path!${RESET}" 1>&2
  exit 1
}

# Setup application variables
MODULE_PATH=$(sed -n -e 's/^.*\(module \)/\1/p' "$PROJECT_PATH/go.mod" | cut -f2- -d' ')
GIT_COMMIT=$(git rev-list -1 HEAD)
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
BUILD_EPOCH=$(date +'%s')
VERSION=$(tr -d '[:space:]' < "$PROJECT_PATH/.version")

# Cleanup previous files and folders
if [ -n "$CLEAN" ]; then
  if [ -n "$VERBOSE" ]; then
    echo -e "${MAGENTA}Cleaning up old bin files...${RESET}"
  fi
  rm "$PROJECT_PATH/bin/osx/$APPLICATION_NAME" &>/dev/null
  rm -r "$PROJECT_PATH/bin/osx" &>/dev/null
  rm "$PROJECT_PATH/bin/windows/$APPLICATION_NAME.exe" &>/dev/null
  rm -r "$PROJECT_PATH/bin/windows" &>/dev/null
  rm "$PROJECT_PATH/bin/linux/$APPLICATION_NAME" &>/dev/null
  rm -r "$PROJECT_PATH/bin/linux" &>/dev/null
  rm "$PROJECT_PATH/bin/$APPLICATION_NAME-$VERSION-osx_64bit.zip" &>/dev/null
  rm "$PROJECT_PATH/bin/$APPLICATION_NAME-$VERSION-windows_64bit.zip" &>/dev/null
  rm "$PROJECT_PATH/bin/$APPLICATION_NAME-$VERSION-linux_64bit.zip" &>/dev/null
fi

# Create new clean directories
if [ -n "$VERBOSE" ]; then
  echo -e "${MAGENTA}Creating new clean bin directories...${RESET}"
fi
mkdir -p "$PROJECT_PATH/bin/osx/"
mkdir -p "$PROJECT_PATH/bin/windows/"
mkdir -p "$PROJECT_PATH/bin/linux/"

# Setup application build time variables
LDFLAGS="-s -w"
LDFLAGS+=" -X $MODULE_PATH/cmd.ApplicationName=$APPLICATION_NAME"
LDFLAGS+=" -X $MODULE_PATH/cmd.BuildBranch=$GIT_BRANCH"
LDFLAGS+=" -X $MODULE_PATH/cmd.BuildRevision=$GIT_COMMIT"
LDFLAGS+=" -X $MODULE_PATH/cmd.BuildVersion=$VERSION"
LDFLAGS+=" -X $MODULE_PATH/cmd.BuildEnv=production"
LDFLAGS+=" -X $MODULE_PATH/cmd.BuildDate=$BUILD_EPOCH"

# Build go binary
echo -e "${MAGENTA}Compiling go binaries...${RESET}"

# Compile for OSX
if [ -n "$OSX" ]; then
  if [ -n "$VERBOSE" ]; then
    echo -e "${MAGENTA}Compiling OSX binary...${RESET}"
  fi
  CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
    -o "$PROJECT_PATH/bin/osx/$APPLICATION_NAME" \
    -ldflags "$LDFLAGS" \
    . || {
    echo -e "${RED} failed to build osx binary!${RESET}" 1>&2
    exit 1
  }
fi

# Compile for Windows
if [ -n "$WINDOWS" ]; then
  if [ -n "$VERBOSE" ]; then
    echo -e "${MAGENTA}Compiling Windows binary...${RESET}"
  fi
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build \
    -o "$PROJECT_PATH/bin/windows/$APPLICATION_NAME.exe" \
    -ldflags "$LDFLAGS" \
    . || {
    echo -e "${RED} failed to build windows binary!${RESET}" 1>&2
    exit 1
  }
fi

# Compile for Linux
if [ -n "$LINUX" ]; then
  if [ -n "$VERBOSE" ]; then
    echo -e "${MAGENTA}Compiling Linux binary...${RESET}"
  fi
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -o "$PROJECT_PATH/bin/linux/$APPLICATION_NAME" \
    -ldflags "$LDFLAGS" \
    . || {
    echo -e "${RED} failed to build linux binary!${RESET}" 1>&2
    exit 1
  }
fi

# Package binaries into ZIP files
if [ -n "$ZIP" ]; then
  if [ -n "$OSX" ]; then
    if [ -n "$VERBOSE" ]; then
      echo -e "${MAGENTA}Packaging OSX binary...${RESET}"
    fi
    zip -j "$PROJECT_PATH/bin/$APPLICATION_NAME-osx_64bit.zip" \
      "$PROJECT_PATH/bin/osx/$APPLICATION_NAME" \
      "$PROJECT_PATH/LICENSE" \
      "$PROJECT_PATH/README.md" \
      "$PROJECT_PATH/CHANGELOG.md"
  fi

  if [ -n "$WINDOWS" ]; then
    if [ -n "$VERBOSE" ]; then
      echo -e "${MAGENTA}Packaging Windows binary...${RESET}"
    fi
    zip -j "$PROJECT_PATH/bin/$APPLICATION_NAME-windows_64bit.zip" \
      "$PROJECT_PATH/bin/windows/$APPLICATION_NAME.exe" \
      "$PROJECT_PATH/LICENSE" \
      "$PROJECT_PATH/README.md" \
      "$PROJECT_PATH/CHANGELOG.md"
  fi

  if [ -n "$LINUX" ]; then
    if [ -n "$VERBOSE" ]; then
      echo -e "${MAGENTA}Packaging Linux binary...${RESET}"
    fi
    zip -j "$PROJECT_PATH/bin/$APPLICATION_NAME-linux_64bit.zip" \
      "$PROJECT_PATH/bin/linux/$APPLICATION_NAME" \
      "$PROJECT_PATH/LICENSE" \
      "$PROJECT_PATH/README.md" \
      "$PROJECT_PATH/CHANGELOG.md"
  fi
fi

echo -e "${MAGENTA}Finished...$RESET"