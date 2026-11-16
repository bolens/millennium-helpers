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
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.1.1/millennium-helpers-v4.1.1-darwin-arm64.tar.gz"
      sha256 "924c8a370ee96597683a0e2ce0dcb26811acacb0797b1157617559cce8ab74cb"
    end
    on_intel do
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.1.1/millennium-helpers-v4.1.1-darwin-amd64.tar.gz"
      sha256 "39d9624fce58a31d203fbfb66264603c43a9023b02ea2243f0337093406b21ff"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.1.1/millennium-helpers-v4.1.1-linux-arm64.tar.gz"
      sha256 "1b4f9ab3f4cb5b438726bd0fbc6aae9d4b0b9980715292f50ba3a4bd029a4ee0"
    end
    on_intel do
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.1.1/millennium-helpers-v4.1.1-linux-amd64.tar.gz"
      sha256 "3711b6c5efa850c4085210fd74b715f831391c6c70c3c104f5a2df996ea31fd8"
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
