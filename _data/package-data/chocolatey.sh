#!/bin/bash

set -o nounset   # disallow usage of unset vars  ( set -u )
set -o errexit   # Exit immediately if a pipeline returns non-zero.  ( set -e )
set -o errtrace  # Allow the above trap be inherited by all functions in the script.  ( set -E )
set -o pipefail  # Return value of a pipeline is the value of the last (rightmost) command to exit with a non-zero status
IFS=$'\n\t'      # Set $IFS to only newline and tab.

cd "$(dirname "$0")/chocolatey"

git clean -ffdX


version=$(cd ../../../ && git tag --sort=-v:refname | grep -P 'v[0-9\.]' | head -1 | cut -c2-)
sed --regexp-extended  -i "s!<version>[0-9\.]+</version>!<version>${version}</version>!g" ffsclient.nuspec


cp "../../../_out/ffsclient_win-386.exe" "/tmp/ffsclient.exe"
zip "tools/ffsclient_32.zip" "/tmp/ffsclient.exe"
rm "/tmp/ffsclient.exe"

cp "../../../_out/ffsclient_win-amd64.exe" "/tmp/ffsclient.exe"
zip "tools/ffsclient_64.zip" "/tmp/ffsclient.exe"
rm "/tmp/ffsclient.exe"


# choco pack

# choco push ffsclient.nupkg --source https://push.chocolatey.org/

