ffsclient - firefox-sync-client
===============================

A commandline-utility to list/view/edit/delete entries in a firefox-sync account.  
Can be used to access bookmarks, passwords, forms, tabs, history or custom data.

[![asciicast](https://asciinema.org/a/533143.svg)](https://asciinema.org/a/533143)

Table of contents
=================
   * [Installation](#installation)
   * [Usage](#usage)
   * [Examples](#example)
      * [Get all bookmarks as json](#get-all-bookmarks-as-json)
      * [Get all bookmarks in netscape format (same as firefox bookmarks.html)](#get-all-bookmarks-in-netscape-format-same-as-firefox-bookmarkshtml)
      * [Get a single bookmark](#get-a-single-bookmark)
      * [Create a new bookmark](#create-a-new-bookmark)
      * [List all passwords](#list-all-passwords)
      * [List deleted passwords](#list-deleted-passwords)
      * [Add a new password](#add-a-new-password)
      * [Delete any record](#delete-any-record)
      * [Query the server without a session file](#query-the-server-without-a-session-file)
      * [Create and read unencrypted records](#create-and-read-unencrypted-records)
   * [Delete a password](#manual)
   * [Request Flowchart](#request-flowchart)
   

Installation
============

The latest binary (windows/linux/macOS/FreeBSD/OpenBSD) can be downloaded from the [github releases page](https://github.com/Mikescher/firefox-sync-client/releases):.

https://github.com/Mikescher/firefox-sync-client/releases/latest

ffsclient does not have any dependencies and can be placed directly in your $PATH (eg /usr/local/bin).

&nbsp;

Alternatively you can use one of the following package manager:
 - [Arch User Repository](https://aur.archlinux.org/packages/ffsclient): `yay -S ffsclient-git` / `yay -S ffsclient-bin`
 - [Homebrew](https://formulae.brew.sh/formula/ffsclient): `brew tap Mikescher/tap && brew install ffsclient`
 - [Chocolatey](https://community.chocolatey.org/packages/ffsclient): `choco install ffsclient`

Usage
=====

Before the first use you have to authenticate your client and create a session.  
Call `ffsclient login {username} {password}` and a session will be created in `~/.config/firefox-sync-client.secret`

After this you can freely use ffsclient (see the full [Manual](#manual) for all commands).  
Common commands are `ffsclient collections` to list all collections in the current account, `ffsclient list {collection}` to list all records in a collection and `ffsclient get {collection} {record-id}` to get a single record.

Almost all commands support different output-formats that can be specified with `--format {fmt}`, available are `text`, `json`, `xml`, `table`, `netscape`, `csv`, `tsv`. If no format is supplied the command uses a default.

You can get an overview of all commands by invoking `ffsclient --help` and a command-specific help with `ffsclient {command} --help`

For some collections (like `bookmarks`, `passwords`, `forms`, `history`, `tabs`) are specific subcommands available.
For example, you can list your bookmarks with `ffsclient bookmarks list`, this is preferable to the general `ffsclient list {collection}` call, because the bookmark-data in the records gets directly parsed and properly displayed.

Example
=======

Here I try to show some common usage patterns:

Get all bookmarks as json  
-------------------------
```
$ ./ffsclient bookmarks list --format json --output bookmarks.json --ignore-schema-errors
$ ./ffsclient bookmarks list --format json --output bookmarks.json --ignore-schema-errors --minimized-json
```
*`--ignore-schema-errors` skips records that are in the bookmarks collection but do not contain valid data*

Get all bookmarks in netscape format (same as firefox bookmarks.html)
---------------------------------------------------------------------
```
$ ./ffsclient bookmarks list --format netscape
```

Get a single bookmark
---------------------
```
$ ./ffsclient get bookmarks "{bookmark_id}" --decoded --format json --pretty-print
```
The `--pretty-print` flag does format the json in the record payload, the surrounding json (from `--format json`) is pretty-printed by default.
If you don't want to have the envelope pretty-printed use the `--minimized-json` flag

Create a new bookmark
---------------------
```
$ ./ffsclient bookmarks create "{title}" "{url}"
$ ./ffsclient bookmarks create "{title}" "{url}" --parent "{parent-record-id}"
$ ./ffsclient bookmarks create "{title}" "{url}" --parent "{parent-record-id}" --position "{index}"
```
By default, bookmarks are created at the top-level and at the last position in the parent folder.

List all passwords
------------------
```
$ ./ffsclient passwords list --ignore-schema-errors
$ ./ffsclient passwords list --show-passwords
```
By default passwords are hidden to prevent accidental leaks.

List deleted passwords
----------------------
```
$ ./ffsclient passwords list --only-deleted
```
Also useful is the `--include-deleted` flag to show both, deleted and normal entries.  
Also works with other list commands

Add a new password
------------------
```
$ ./ffsclient passwords create "{url}" "{username}" "{password}"
```

Delete a password
------------------
```
$ ./ffsclient passwords delete "{url}"
$ ./ffsclient passwords delete "{record-id}"
```

Delete any record
------------------
```
$ ./ffsclient delete "{record-id}"
$ ./ffsclient delete "{record-id}" --hard
```
*By default, the sync protocol needs tombstones. This means deleted records still exist, without their payload and wit a deleted:true flag*  
*If the `--hard` flag is supplied, the record is instead completely deleted from the server*

Query the server without a session file
---------------------------------------
```
$ ./ffsclient collections --auth-login-email "{username}" --auth-login-password "{password}"
```
This does not use the normal session file an creates a completely new session for this command only.  
This is genrally **not** recommended to do. Your request will look like a new client to the server, it can happen that you have to allow it via email and it is also much more inefficient.  
If you don't want a session file in your home folder use `--sessionfile` to specify a more secure location

Create and read unencrypted records
-----------------------------------
```
$ ./ffsclient create "{collection}" "{id}" --raw "hello world"
$ ./ffsclient get "{collection}" "{id}" --raw --format json
$ ./ffsclient get "{collection}" "{id}" --raw --format text --data-only
```
Normally records have an encrypted payload that needs to be decrypted before it can be read (via the `--decrypted` flag in `ffsclient get`).  
But you can also directly write data in the payload field.  
The `--raw` flag in `ffsclient create` skips the normal encryption step and the `--raw` flag in `ffsclient get` skips the decryption.  
This is only recommended for custom collections, you should never write invalid data in one of the default collections (e.g. `bookmarks`, `passwords`, etc)

Manual
======

*(copied from v1.4.0)*

*If I forgot to update the README you can always get the current version of the help with `./ffsclient --help`*

```
firefox-sync-client.

# (Use `ffsclient <command> --help` for more detailed info)

Basic Usage:
  ffsclient login <login> <password>                               Login to FF-Sync account, uses ~/.config as default session location
            [--device-name=<name>]                                   # Send your device-name to identify the session later
            [--device-type=<type>]                                   # Send your device-type to identify the session later
            [--otp=<value>]                                          # A valid TOTP token, in case one is needed for the login
  ffsclient refresh [--force]                                      Refresh the current session token (via OAuth RefreshToken)
  ffsclient check-session                                          Verify that the current session is valid
  ffsclient collections                                            List all available collections
            [--usage]                                                # Include usage (storage space)
  ffsclient quota                                                  Query the storage quota of the current user
  ffsclient list <collection>                                      Get a all records in a collection (use --format to define the format)
            (--raw | --decoded | --ids)                              # Return raw data, decoded payload, or only IDs
            [--after <rfc3339>]                                      # Return only fields updated after this date
            [--sort <sort>]                                          # Sort the result by (newest|index|oldest)
            [--limit <n>]                                            # Return max <n> elements
            [--offset <o>]                                           # Skip the first <n> elements
            [--pretty-print | --pp]                                  # Pretty-Print json in decoded data / payload (if possible)
  ffsclient get <collection> <record-id>                           Get a single record
            (--raw | --decoded)                                      # Return raw data or decoded payload
            [--pretty-print | --pp]                                  # Pretty-Print json in decoded data / payload (if possible)
            [--data-only]                                            # Only return the payload
  ffsclient delete <collection> <record-id> [--hard]               Delete the specified record
  ffsclient delete <collection>                                    Delete all the records in a collection
  ffsclient delete-all --force                                     Delete all (!) records in the server
  ffsclient create <collection> <record-id>                        Insert a new record
            (--raw <r> | --data <d> | --raw-stdin | --data-stdin)    # The new data
  ffsclient update <collection> <record-id>                        Update an existing record
            (--raw <r> | --data <d> | --raw-stdin | --data-stdin)    # The new data
            [--create]                                               # Create a new record if the specified record-id does not exist
  ffsclient meta                                                   Get storage metadata
  ffsclient <sub> --help                                           Output specific help for a single subcommand

Usage:
  ffsclient bookmarks list                                         List bookmarks (use --format to define the format)
            [--ignore-schema-errors]                                 # Skip records that cannot be decoded into a bookmark schema
            [--after <rfc3339>]                                      # Return only fields updated after this date
            [--sort <sort>]                                          # Sort the result by (newest|index|oldest)
            [--limit <n>]                                            # Return max <n> elements
            [--offset <o>]                                           # Skip the first <n> elements
            [--include-deleted]                                      # Show deleted entries
            [--only-deleted]                                         # Show only deleted entries
            [--type <folder|separator|bookmark|...>]                 # Show only entries with the specified type
            [--parent <id>]                                          # Show only entries with the specified parent (by record-id), can be specified multiple times
            [--linear                                                # Do not output the folder hierachy
  ffsclient bookmarks delete <id>                                  Delete the specified bookmark
  ffsclient bookmarks create bookmark <title> <url>                Insert a new bookmark
            [--description <desc>]                                   # Specify the bookmark description
            [--load-in-sidebar]                                      # If specified the `LoadInSidebar` field is set to true (default is false)
            [--tag <tag>]                                            # Add a tag to the bookmark, specify multiple times to add multiple tags
            [--keyword <kw>]                                         # Specify the keyword (to activate the bookmark from the location bar)
            [--parent <id>]                                          # Specify the ID of the parent folder (if not specified the entry lives under `unfiled`)
            [--position=<idx>]                                       # The position of the entry in the parent (0 = first, default is last). Can use negative indizes.
  ffsclient bookmarks create folder <title>                        Insert a new bookmark-folder
            [--parent <id>]                                          # Specify the ID of the parent folder (if not specified the entry lives under `unfiled`)
            [--position=<idx>]                                       # The position of the entry in the parent (0 = first, default is last). Can use negative indizes.
  ffsclient bookmarks create separator                             Insert a new bookmark-separator
            [--parent <id>]                                          # Specify the ID of the parent folder (if not specified the entry lives under `unfiled`)
            [--position=<idx>]                                       # The position of the entry in the parent (0 = first, default is last). Can use negative indizes.
  ffsclient bookmarks update <id>                                  Partially update a bookmark
            [--title <title>]                                        # Change the bookmark title
            [--url <url>]                                            # Change the URL
            [--description <desc>]                                   # Change the bookmark description
            [--load-in-sidebar <true|false>]                         # Set the `LoadInSidebar` field
            [--tag <tag>]                                            # Change the tags, specify multiple times to set multiple tags
            [--keyword <kw>]                                         # Specify the keyword (to activate the bookmark from the location bar)
            [--position=<idx>]                                       # Change the position of the entry in the parent (0 = first). Can use negative indizes.
  ffsclient passwords list                                         List passwords
            [--show-passwords]                                       # Show the actual passwords
            [--ignore-schema-errors]                                 # Skip records that cannot be decoded into a password schema
            [--after <rfc3339>]                                      # Return only fields updated after this date
            [--sort <sort>]                                          # Sort the result by (newest|index|oldest)
            [--limit <n>]                                            # Return max <n> elements
            [--offset <o>]                                           # Skip the first <n> elements
            [--include-deleted]                                      # Show deleted entries
            [--only-deleted]                                         # Show only deleted entries
  ffsclient passwords delete <host|id> [--hard]                    Delete a single password
            [--is-host | --is-exact-host | --is-id]                  # Specify that the supplied argument is a host / record-id (otherwise both is possible)
  ffsclient passwords create <host> <username> <password>          Insert a new password
            [--form-submit-url <url>]                                # Specify the submission URL (GET/POST url set by <form>)
            [--http-realm <realm>]                                   # Specify the HTTP Realm (HTTP Realm for which the login is valid)
            [--username-field <name>]                                # Specify the Username field (HTML field name of the username)
            [--password-field <name>]                                # Specify the Password field (HTML field name of the password)
  ffsclient passwords update <host|id>                             Update an existing password
            [--is-host | --is-exact-host | --is-id]                  # Specify that the supplied argument is a host / record-id (otherwise both is possible)
            [--host <url>]                                           # Update the host field
            [--username <user>]                                      # Update the username
            [--password <pass>]                                      # Update the password
            [--form-submit-url <url>]                                # Update the submission URL (GET/POST url set by <form>)
            [--http-realm <realm>]                                   # Update the HTTP Realm (HTTP Realm for which the login is valid)
            [--username-field <name>]                                # Update the Username field (HTML field name of the username)
            [--password-field <name>]                                # Update the Password field (HTML field name of the password)
  ffsclient passwords get <host|id>                                Insert a new password
            [--is-host | --is-exact-host | --is-id]                  # Specify that the supplied argument is a host / record-id (otherwise both is possible)
  ffsclient forms list                                             List form autocomplete suggestions
            [--name <n>]                                             # Show only entries with the specified name
            [--ignore-schema-errors]                                 # Skip records that cannot be decoded into a form schema
            [--after <rfc3339>]                                      # Return only fields updated after this date
            [--sort <sort>]                                          # Sort the result by (newest|index|oldest)
            [--limit <n>]                                            # Return max <n> elements
            [--offset <o>]                                           # Skip the first <n> elements
            [--include-deleted]                                      # Show deleted entries
            [--only-deleted]                                         # Show only deleted entries
  ffsclient forms get <name> [--ignore-case]                       Get all HTML-Form autocomplete suggestions for this name
  ffsclient forms create <name> <value>                            Adds a new HTML-Form autocomplete suggestions
  ffsclient forms delete <id> [--hard]                             Delete the specified HTML-Form autocomplete suggestion
  ffsclient history list                                           List form history entries
            [--ignore-schema-errors]                                 # Skip records that cannot be decoded into a history schema
            [--after <rfc3339>]                                      # Return only fields after this date
            [--sort <sort>]                                          # Sort the result by (newest|index|oldest)
            [--limit <n>]                                            # Return max <n> elements
            [--offset <o>]                                           # Skip the first <n> elements
            [--include-deleted]                                      # Show deleted entries
            [--only-deleted]                                         # Show only deleted entries
  ffsclient history delete <id> [--hard]                           Delete the specified history entry
  ffsclient tabs list                                              List synchronized tabs
            [--client <n>]                                           # Show only entries from the specified client (must be a valid client-id)
            [--ignore-schema-errors]                                 # Skip records that cannot be decoded into a tab schema
            [--limit <n>]                                            # Return max <n> elements (clients)
            [--offset <o>]                                           # Skip the first <n> elements (clients)
            [--include-deleted]                                      # Show deleted entries
            [--only-deleted]                                         # Show only deleted entries

Hint:
  # If you need to supply a record-id / collection that starts with an minus, use the --!arg=... syntax
  #     e.g.: `ffsclient get bookmarks --!arg=-udhG86-JgpUx --decoded`
  # Also if you need to supply a argument that starts with an - use the --arg=value syntax
  #     e.g.: `ffsclient bookmarks add Test "https://example.org" --parent toolbar --position=-3`

Common Options:
  -h, --help                                                       Show this screen.
  --version                                                        Show version.
  -v, --verbose                                                    Output more intermediate information
  -q, --quiet                                                      Do not print anything
  -f <fmt>, --format <fmt>                                         Specify the output format (not all subcommands support all output-formats)
                                                                     # - 'text'
                                                                     # - 'json'
                                                                     # - 'netscape'   (default firefox bookmarks format)
                                                                     # - 'xml'
                                                                     # - 'table'
                                                                     # - 'csv'
                                                                     # - 'tsv'
  --table-truncate                                                 Truncate columns of table-format to fit terminal width (needs -f table)
  --no-table-truncate                                              Disable truncation of columns in table-format output
  --table-columns <col-list>                                       Limit displayed columns of table-format output (comma-seperated list of headers)
  --csv-filter                                                     Only print specified columns in csv/tsv output (comma-seperated index list) (needs -f csv/tsv)
  --auth-server <url>                                              Specify the (authentication) server-url
  --token-server <url>                                             Specify the (token) server-url
  --request-retry-delay-certerr <sec>                              Retry delay for requests that had a certificate error (default: 5 sec)
  --request-retry-delay-floodcontrol <sec>                         Retry delay for requests that were throttled by the server (default: 15 sec)
  --request-retry-delay-servererr <sec>                            Retry delay for requests that failed due to server errors (default: 1 sec)
  --request-retry-max <num>                                        Max request retries (default: 5)
  --request-timeout <sec>                                          Timeout for API request (default 10 sec)
  --request-ignore-certerr                                         Ignore certificate errors (do not verify ssl)
  --color                                                          Enforce colored output
  --no-color                                                       Disable colored output
  --timezone <tz>                                                  Specify the output timezone
                                                                     # Can be either:
                                                                     #   - UTC
                                                                     #   - Local (default)
                                                                     #   - IANA Time Zone, e.g. 'America/New_York'
  --timeformat <url>                                               Specify the output timeformat (golang syntax)
  -o <f>, --output <f>                                             Write the output to a file
  --sessionfile <cfg>                                              Specify the location of the saved session
  --auth-login-email <email>                                       Login with the sync server without using the saved session (enforces a new, temporary session)
  --auth-login-password <pw>                                       Login with the sync server without using the saved session (enforces a new, temporary session)
  --no-autosave-session                                            Do not update the sessionfile if the session was auto-refreshed
  --force-refresh-session                                          Always auto-refresh the session, even if its not expired
  --no-xml-declaration                                             Do not print the xml declaration when using `--format xml`
  --minimized-json                                                 Do not indent (pretty-print) json output when using `--format json`

Exit Codes:
  0             Program exited successfully
  60            Program existed with an (unspecified) error
  61            Program crashed
  62            Program called without arguments
  63            Failed to parse commandline arguments
  64            Command needs a valid session/session-file and none was found
  65            The current subcommand does not support the specified output format
  66            Record with this ID not found

  81            (check-session): The session is not valid
  82            (passwords): No matching password found
  83            (create-bookmarks): Parent record is not a folder
  84            (create-bookmarks): The position in the parent would be out of bounds
  85            (update-bookmarks): One of the specified fields is not valid on the record type
```




Request Flowchart
=================

![FFSync-Flowchart](_data/readme-data/api-flow.svg)
