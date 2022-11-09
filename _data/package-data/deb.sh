#!/bin/bash

# https://linuxconfig.org/easy-way-to-create-a-debian-package-and-local-package-repository
# https://www.debian.org/doc/debian-policy/ch-controlfields.html

set -o nounset   # disallow usage of unset vars  ( set -u )
set -o errexit   # Exit immediately if a pipeline returns non-zero.  ( set -e )
set -o errtrace  # Allow the above trap be inherited by all functions in the script.  ( set -E )
set -o pipefail  # Return value of a pipeline is the value of the last (rightmost) command to exit with a non-zero status
IFS=$'\n\t'      # Set $IFS to only newline and tab.


cd "$(dirname "$0")/../../"

version="$(git tag --sort=-v:refname | grep -P 'v[0-9\.]' | head -1 | cut -c2-)"
fsize="$(wc -c _out/ffsclient_linux-amd64  | awk '{print int($1 / 1024)*1024}')"

mkdir _out/deb_amd64
mkdir _out/deb_amd64/ffsclient
mkdir _out/deb_amd64/ffsclient/DEBIAN

cp _data/package-data/deb/control _out/deb_amd64/ffsclient/DEBIAN/control

sed --regexp-extended  -i "s/<<version>>/${version}/g" _out/deb_amd64/ffsclient/DEBIAN/control
sed --regexp-extended  -i "s/<<arch>>/amd64/g"         _out/deb_amd64/ffsclient/DEBIAN/control
sed --regexp-extended  -i "s/<<fsize>>/${fsize}/g"     _out/deb_amd64/ffsclient/DEBIAN/control

mkdir _out/deb_amd64/ffsclient/usr
mkdir _out/deb_amd64/ffsclient/usr/bin

cp _out/ffsclient_linux-amd64 _out/deb_amd64/ffsclient/usr/bin/

docker run --rm --volume "$(pwd)/_out/deb_amd64/:/deb/" debian dpkg-deb --build /deb/ffsclient

cp -v _out/deb_amd64/ffsclient.deb _out/ffsclient_amd64.deb