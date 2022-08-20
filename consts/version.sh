#!/bin/bash

sed -i 's/const FFSCLIENT_VERSION = ".*"/const FFSCLIENT_VERSION = "'$(git describe --tags | sed "s/v//")'"/' "version.go"
