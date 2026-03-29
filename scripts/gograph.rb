class Gograph < Formula
  desc "Pure Go embedded graph database"
  homepage "https://github.com/DotNetAge/gograph"
  version "v0.2.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DotNetAge/gograph/releases/download/v0.2.0/gograph-v0.2.0-darwin-arm64.tar.gz"
      sha256 "53e676ecc084de771e5deec75e96962bd11654b5085302ba00ee5585d2e50adb"
    else
      url "https://github.com/DotNetAge/gograph/releases/download/v0.2.0/gograph-v0.2.0-darwin-amd64.tar.gz"
      sha256 "3e4771b1fc148e68ec1d3ec3eda1c5a5bb1b086de9d61292086ff7be2de06b48"
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/DotNetAge/gograph/releases/download/v0.2.0/gograph-v0.2.0-linux-arm64.tar.gz"
      sha256 "bacd8831cd55093e257a79f1fc82dc5b906e8049d50d20ea86de0635e7a6de76"
    else
      url "https://github.com/DotNetAge/gograph/releases/download/v0.2.0/gograph-v0.2.0-linux-amd64.tar.gz"
      sha256 "d93665425264cfd25319d5cb8f1cd67c76fae55011e1c07060d18f7cab475d6d"
    end
  end

  def install
    bin.install "gograph"
  end

  test do
    system "#{bin}/gograph", "query", "MATCH (n) RETURN n"
  end
end
