#!/bin/bash

echo ""
echo "APIKey: $1"
echo ""

echo ""
echo "File:   /root/ffsclient/ffsclient.nupkg"
echo ""

echo ""
stat /root/ffsclient/*.nupkg
echo ""

choco push /root/ffsclient/*.nupkg --api-key "$1" --source https://push.chocolatey.org/

