class Portman < Formula
  desc "Fast, colorful port usage analyzer with TUI process management"
  homepage "https://github.com/heartwilltell/portman"
  url "https://github.com/heartwilltell/portman/archive/d8a3a8b554b1885d0704b62f7373d837dbd646d3.tar.gz"
  sha256 "78889b48a587af9c76a3841cd6f70fbef4a898d2d87fa0436eedb5d60ef1a8f2"
  license "MIT"
  version "0.1.0"
  head "https://github.com/heartwilltell/portman.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X main.Version=#{version}
    ]

    system "go", "build", *std_go_args(ldflags: ldflags, output: bin/"portman")
  end

  test do
    output = shell_output("#{bin}/portman -help")
    assert_match "Portman - Port Usage Analyzer", output
  end
end
