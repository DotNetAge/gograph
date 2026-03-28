class Gograph < Formula
  desc "Pure Go embedded graph database"
  homepage "https://github.com/DotNetAge/gograph"
  url "https://github.com/DotNetAge/gograph/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "71db0954df271fe71a6ff4d57c64625fb4632ab796d98405bd9a27afc9c5234f"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", "-ldflags", "-X main.Version=#{version}", "-o", bin/"gog", "./cmd/gograph"
  end

  test do
    system "#{bin}/gog", "query", "MATCH (n) RETURN n"
  end
end
