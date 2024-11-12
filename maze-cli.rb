class maze-cli < Formula
    desc "A CLI application for maze multicloud platform"
    homepage "https://github.ibm.com/MAZE-MultiCloud/homebrew-maze-cli"
    url "https://github.ibm.com/MAZE-MultiCloud/homebrew-maze-cli/releases/download/v1.0.0/maze-macos"
    sha256 "YOUR_BINARY_SHA256" # Replace with the actual SHA256 checksum of the binary
    version "1.0.0" # Replace with your app version
  
    def install
      bin.install "maze-cli"
    end
  
    test do
      system "#{bin}/maze-cli", "--help" # A simple test to verify installation
    end
  end
  