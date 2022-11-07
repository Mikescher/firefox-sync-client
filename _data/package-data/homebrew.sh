#!/bin/bash

set -o nounset   # disallow usage of unset vars  ( set -u )
set -o errexit   # Exit immediately if a pipeline returns non-zero.  ( set -e )
set -o errtrace  # Allow the above trap be inherited by all functions in the script.  ( set -E )
set -o pipefail  # Return value of a pipeline is the value of the last (rightmost) command to exit with a non-zero status
IFS=$'\n\t'      # Set $IFS to only newline and tab.

set -o functrace

cd "$(dirname "$0")/homebrew"

cp ffsclient.rb ffsclient_patch.rb


version="$(cd ../../../ && git tag --sort=-v:refname | grep -P 'v[0-9\.]' | head -1 | cut -c2-)"
cs0="$(cd ../../../ && sha256sum _out/ffsclient_linux-amd64 | cut -d ' ' -f 1)"

echo "Version: ${version} (${cs0})"

sed --regexp-extended  -i "s/<<version>>/${version}/g"  ffsclient_patch.rb
sed --regexp-extended  -i "s/<<shahash>>/${cs0}/g"      ffsclient_patch.rb

cd ../../../
git clone https://github.com/Mikescher/homebrew-tap.git _out/homebrew-tap

cp "_data/package-data/homebrew/ffsclient_patch.rb" _out/homebrew-tap/ffsclient.rb
rm "_data/package-data/homebrew/ffsclient_patch.rb"


cd _out/homebrew-tap/

git add ffsclient.rb

git commit -m "ffsclient v${version}"


# git push manually (!)
