class Gograph < Formula
  desc "Pure Go embedded graph database"
  homepage "https://github.com/DotNetAge/gograph"
  version "v0.2.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/DotNetAge/gograph/releases/download/v0.2.2/gograph-v0.2.2-darwin-arm64.tar.gz"
      sha256 "3fc84325ae1197394ef8d70d9d38d6c018d8d01fd2a40d2fc0b4d3da236a5fb5"
    else
      url "https://github.com/DotNetAge/gograph/releases/download/v0.2.2/gograph-v0.2.2-darwin-amd64.tar.gz"
      sha256 "34e42cd431fa57a4658780f53fef49cbf5cceff894973d72f8673c53f319d094"
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/DotNetAge/gograph/releases/download/v0.2.2/gograph-v0.2.2-linux-arm64.tar.gz"
      sha256 "ff14d7d80513be1338fbaaedbcd19efddbdfe80b3929549b072cfd89b8e18707"
    else
      url "https://github.com/DotNetAge/gograph/releases/download/v0.2.2/gograph-v0.2.2-linux-amd64.tar.gz"
      sha256 "4032f31b68584a5b60723748a3899bca478e5eff285e38b0664747715a865866"
    end
  end

  def install
    bin.install "gograph"
  end

  test do
    system "#{bin}/gograph", "query", "MATCH (n) RETURN n"
  end
end
