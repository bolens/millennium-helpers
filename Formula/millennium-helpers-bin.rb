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
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.1.0/millennium-helpers-v4.1.0-darwin-arm64.tar.gz"
      sha256 "5a97666211c9f5577ff70bd36822858f2dac934a689d2ebad6df5ede81a3c7c8"
    end
    on_intel do
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.1.0/millennium-helpers-v4.1.0-darwin-amd64.tar.gz"
      sha256 "6c3a17aa913cf4c284e7bdd05704d46f39e78ba2cfe13e1f8f0b09b26c5a9dfc"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.1.0/millennium-helpers-v4.1.0-linux-arm64.tar.gz"
      sha256 "2cfdfc26728de21d63008dd34e0976322f2a6ea2aed6d14ac3f0ee71967b2403"
    end
    on_intel do
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.1.0/millennium-helpers-v4.1.0-linux-amd64.tar.gz"
      sha256 "9f03e57cc5b7733df5efba69f889739e3099a318fa92db340f95262429e235b9"
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

    %w[LICENSE THIRD_PARTY_LICENSES.txt THIRD_PARTY_NOTICES.md].each do |notice|
      (lib/"millennium-helpers").install notice if File.exist?(notice)
    end
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
