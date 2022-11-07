class FirefoxSyncClient < Formula

  desc     " A cli for firefox-sync (firefox bookmarks, passwords, account, ...) "
  homepage "https://github.com/Mikescher/firefox-sync-client"
  url      "https://github.com/Mikescher/firefox-sync-client/releases/download/v<<version>>/ffsclient_macos-amd64"
  sha256   "<<shahash>>"

  def install
    bin.install "firefox-sync-client"
  end

  test do
    assert true
  end

end