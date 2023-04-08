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

TAG_DELIMITER="/"

SPLIT_TAG=(${VERSION_TAG//"$TAG_DELIMITER"/ })
NORMALIZED_TAG=${SPLIT_TAG[2]}

if [ "${#SPLIT_TAG[@]}" -eq 1 ];
then
  NORMALIZED_TAG="$VERSION_TAG"
fi

GO_PROXY_URL=https://proxy.golang.org/github.com/alexandria-oss/"$MODULE_NAME"/@v/"$NORMALIZED_TAG".info
echo "forcing package publishing using URL: $GO_PROXY_URL"
curl "$GO_PROXY_URL"
