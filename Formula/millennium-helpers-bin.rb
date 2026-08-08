class MillenniumHelpersBin < Formula
  desc "Prebuilt CLI and helpers for managing Millennium Steam mods"
  homepage "https://github.com/bolens/millennium-helpers"
  license "MIT"

  depends_on "bash"
  depends_on "curl"
  depends_on "jq"
  depends_on "python"
  depends_on "unzip"

  on_macos do
    on_arm do
      url "https://github.com/bolens/millennium-helpers/releases/download/v3.2.1/millennium-helpers-v3.2.1-darwin-arm64.tar.gz"
      sha256 "ba9feca700cefbded45581609d025f0578da55f4e9a6d0ea8dd4454c6c9589ea"
    end
    on_intel do
      url "https://github.com/bolens/millennium-helpers/releases/download/v3.2.1/millennium-helpers-v3.2.1-darwin-amd64.tar.gz"
      sha256 "ca83283b1e1a13106f0c6eb555c8d92bc684d91c6ae3ea531a80c58838d3763b"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/bolens/millennium-helpers/releases/download/v3.2.1/millennium-helpers-v3.2.1-linux-arm64.tar.gz"
      sha256 "c6235eb62e568723b272808d6be00f9bee0cdbedf890199892c9605bf232f65c"
    end
    on_intel do
      url "https://github.com/bolens/millennium-helpers/releases/download/v3.2.1/millennium-helpers-v3.2.1-linux-amd64.tar.gz"
      sha256 "fe220070208776338e6da005388340c9aa1b3a0b3240cb8aa4c1948a1976dedf"
    end
  end

  conflicts_with "millennium-helpers", because: "both install the millennium helper tools"

  def install
    odie "Release archive missing bin/millennium (Go dispatcher required)" unless (buildpath/"bin/millennium").exist?
    bin.install "bin/millennium"
    bash_completion.install "completions/bash/millennium-helpers" => "millennium-helpers"
    ln_sf "millennium-helpers", bash_completion/"millennium"

    zsh_completion.install "completions/zsh/_millennium-helpers" => "_millennium-helpers"
    ln_sf "_millennium-helpers", zsh_completion/"_millennium"

    fish_completion.install "completions/fish/millennium.fish"
    (share/"nushell/completions").install "completions/nushell/millennium-helpers.nu"
    man1.install Dir["man/*.1"]
    (lib/"millennium-helpers").install "VERSION"

    license_md = "third_party/MILLENNIUM-LICENSE.md"
    (lib/"millennium-helpers").install license_md if File.exist?(license_md)
  end

  def caveats
    <<~EOS
      This formula installs the published OS/arch release tarball.
      For a from-source build with `go`, use: millennium-helpers
    EOS
  end

  test do
    system "#{bin}/millennium", "diag", "--help"
    assert_path_exists lib/"millennium-helpers/VERSION"
  end
end
