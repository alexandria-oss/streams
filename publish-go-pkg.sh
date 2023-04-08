#!/bin/bash

while getopts ":v:m:" opt; do
  case $opt in
    v) VERSION_TAG=$OPTARG;;
      \?)
        echo "Invalid option: -$OPTARG" >&2
        exit 1
        ;;
    m) MODULE_NAME=$OPTARG;;
      \?)
        echo "Invalid option: -$OPTARG" >&2
        exit 1
        ;;
    :)
      echo "Option -$OPTARG requires an argument." >&2
      exit 1
      ;;
  esac
done

MAJOR_VERSION=""
VERSION_PREFIX=""

if [ "${#VERSION_TAG}" -ge 2 ]
then
  VERSION_PREFIX="${VERSION_TAG:0:2}"
fi

if [ "$VERSION_PREFIX" != "" ] && [ "$VERSION_PREFIX" != "v0" ] && [ "$VERSION_PREFIX" != "v1" ]
then
  MAJOR_VERSION="/$VERSION_PREFIX"
fi

GO_PROXY_URL=https://proxy.golang.org/github.com/alexandria-oss/"$MODULE_NAME""$MAJOR_VERSION"/@v/"$VERSION_TAG".info
echo "forcing package publishing using URL: $GO_PROXY_URL"
curl "$GO_PROXY_URL"
