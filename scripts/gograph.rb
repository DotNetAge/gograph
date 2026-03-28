class Gograph < Formula
  desc "Pure Go embedded graph database"
  homepage "https://github.com/DotNetAge/gograph"
  version "v0.1.0-4-g3f04068"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DotNetAge/gograph/releases/download/vv0.1.0-4-g3f04068/gograph-v0.1.0-4-g3f04068-darwin-arm64.tar.gz"
      sha256 "bf0e8b3cea6154b2228b3fbb0020861f1ce74a23ec9a43d3277335ee0bb4be0f"
    else
      url "https://github.com/DotNetAge/gograph/releases/download/vv0.1.0-4-g3f04068/gograph-v0.1.0-4-g3f04068-darwin-amd64.tar.gz"
      sha256 "15baaa4c86a27de182cf79194b60a13f7788f7ac961f8ef2b8ebb4eee1a1bb92"
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/DotNetAge/gograph/releases/download/vv0.1.0-4-g3f04068/gograph-v0.1.0-4-g3f04068-linux-arm64.tar.gz"
      sha256 "bc3ddd14153f5321a8b1ee05c03f65b009db698424f4e386cf93e881619cfb7f"
    else
      url "https://github.com/DotNetAge/gograph/releases/download/vv0.1.0-4-g3f04068/gograph-v0.1.0-4-g3f04068-linux-amd64.tar.gz"
      sha256 "3478c94eb828e85fd931af6d7e5cbb038c085c30c447f9664c6b2845d8dec3b7"
    end
  end

  def install
    bin.install "gograph"
  end

  test do
    system "#{bin}/gograph", "query", "MATCH (n) RETURN n"
  end
end
