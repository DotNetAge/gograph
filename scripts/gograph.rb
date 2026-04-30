class Gograph < Formula
  desc "Pure Go embedded graph database"
  homepage "https://github.com/DotNetAge/gograph"
  version "v0.2.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DotNetAge/gograph/releases/download/v0.2.3/gograph-v0.2.3-darwin-arm64.tar.gz"
      sha256 "921147676802656b1cb3770a0a6261edcc890de1ba274b1e5222abfa9819a16c"
    else
      url "https://github.com/DotNetAge/gograph/releases/download/v0.2.3/gograph-v0.2.3-darwin-amd64.tar.gz"
      sha256 "d64b8144e111183704048928ca1415d0df193235f39fdd58a7681b7f71630352"
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/DotNetAge/gograph/releases/download/v0.2.3/gograph-v0.2.3-linux-arm64.tar.gz"
      sha256 "6aedfba9bd57a298c7dafd69deed27a203381d320346f8d6fb067f63e5471471"
    else
      url "https://github.com/DotNetAge/gograph/releases/download/v0.2.3/gograph-v0.2.3-linux-amd64.tar.gz"
      sha256 "3444b8763e39a582fa03c208491dd1aba2a362eb1418d3e1076519c3f972ed5c"
    end
  end

  def install
    bin.install "gograph"
  end

  test do
    system "#{bin}/gograph", "query", "MATCH (n) RETURN n"
  end
end
