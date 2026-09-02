class Arca < Formula
  desc "Interactive project setup CLI"
  homepage "https://github.com/ishiguro-junya/arca"
  version "0.0.0"
  url "https://github.com/ishiguro-junya/arca/releases/download/v#{version}/arca_darwin_arm64.tar.gz"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"

  depends_on arch: :arm64
  depends_on :macos

  def install
    bin.install "arca"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/arca version")
  end
end
