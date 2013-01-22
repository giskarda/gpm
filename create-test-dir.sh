#!/bin/bash

set -e

mkdir -p /tmp/test/devel

for file in somethingMore.rpm nothingMore.rpm somethingLess.deb nothingLess.rpm somethingEqual.gz nothingEqual.rpm; do
    touch /tmp/test/devel/$file
done
